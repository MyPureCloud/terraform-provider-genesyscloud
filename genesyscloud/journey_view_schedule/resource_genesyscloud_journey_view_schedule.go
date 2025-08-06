package journey_view_schedule

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
)

/*
The resource_genesyscloud_journey_view_schedule.go contains all the methods that perform the core logic for a resource.
*/

// getAllJourneyViewSchedule retrieves all the journey view schedules via Terraform in the Genesys Cloud and is used for the exporter
func getAllJourneyViewSchedule(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getJourneyViewScheduleProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	journeyViewSchedules, resp, err := proxy.getAllJourneyViewSchedule(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get journey view schedules error: %v", err), resp)
	}

	if journeyViewSchedules == nil || len(*journeyViewSchedules) == 0 {
		return resources, nil
	}

	for _, journeyViewSchedule := range *journeyViewSchedules {
		resources[*journeyViewSchedule.Id] = &resourceExporter.ResourceMeta{BlockLabel: *journeyViewSchedule.Id}
	}

	return resources, nil
}

// createJourneyViewSchedule is used by the journey_view_schedule resource to create Genesys cloud journey view schedule
func createJourneyViewSchedule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	gp := getJourneyViewScheduleProxy(sdkConfig)

	journeyViewId := d.Get("journey_view_id").(string)
	journeyViewSchedule := makeJourneyViewScheduleFromSchema(d)

	log.Printf("Creating schedule for journeyView id: %s", journeyViewId)
	journeyViewSchedule, resp, err := gp.createJourneyViewSchedule(ctx, journeyViewId, journeyViewSchedule)

	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create schedule for journeyView id %s: %s", journeyViewId, err), resp)
	}

	// Currently the create API returns 201 with an empty response body. So the schedule id is nil
	// The schedule entity id is identical to the journey view id, using this to set resource Id.
	// A journey view and its schedule is a 1-1 map
	if journeyViewSchedule.Id == nil {
		d.SetId(journeyViewId)
	} else {
		d.SetId(*journeyViewSchedule.Id)
	}

	log.Printf("Created schedule for journeyView id: %s", journeyViewId)
	return readJourneyViewSchedule(ctx, d, meta)
}

// readJourneyViewSchedule is used by the journey_view_schedule resource to read an journey view schedule from genesys cloud
func readJourneyViewSchedule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	journeyViewId := d.Get("journey_view_id").(string)
	if journeyViewId == "" {
		journeyViewId = d.Id()
	}

	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceJourneyViewSchedule(), constants.ConsistencyChecks(), ResourceType)
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	gp := getJourneyViewScheduleProxy(sdkConfig)

	log.Printf("Reading schedule for journeyView id: %s", journeyViewId)

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		journeyViewSchedule, resp, err := gp.getJourneyViewScheduleByViewId(ctx, journeyViewId)
		if err != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to get schedule for journeyView id %s | error: %s", journeyViewId, err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to get schedule for journeyView id %s | error: %s", journeyViewId, err), resp))
		}

		resourcedata.SetNillableValue(d, "frequency", journeyViewSchedule.Frequency)
		resourcedata.SetNillableValue(d, "journey_view_id", journeyViewSchedule.Id)

		log.Printf("Read schedule for journeyView id %s", journeyViewId)
		return cc.CheckState(d)
	})
}

// updateJourneyViewSchedule is used by the journey_view_schedule resource to update a journey view schedule in Genesys Cloud
func updateJourneyViewSchedule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getJourneyViewScheduleProxy(sdkConfig)

	journeyViewId := d.Get("journey_view_id").(string)
	journeyViewSchedule := makeJourneyViewScheduleFromSchema(d)

	log.Printf("Updating schedule for journeyView id %s ", journeyViewId)

	journeyViewSchedule, resp, err := proxy.updateJourneyViewSchedule(ctx, journeyViewId, journeyViewSchedule)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update schedule for journeyView id %s: %s", journeyViewId, err), resp)
	}

	log.Printf("Updated schedule for journeyView id %s", journeyViewId)
	return readJourneyViewSchedule(ctx, d, meta)
}

// deleteJourneyViewSchedule is used by the journey_view_schedule resource to delete a journey view schedule from Genesys cloud
func deleteJourneyViewSchedule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	journeyViewId := d.Get("journey_view_id").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getJourneyViewScheduleProxy(sdkConfig)

	log.Printf("Deleting schedule for journeyView id: %s", journeyViewId)

	resp, err := proxy.deleteJourneyViewSchedule(ctx, journeyViewId)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete schedule for journeyView id %s: %s", journeyViewId, err), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getJourneyViewScheduleByViewId(ctx, journeyViewId)

		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted schedule for journeyView id %s", journeyViewId)
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting schedule for journeyView id %s: %s", journeyViewId, err), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Schedule for journeyView id %s still exists", journeyViewId), resp))
	})
}

func makeJourneyViewScheduleFromSchema(d *schema.ResourceData) *platformclientv2.Journeyviewschedule {
	frequency := d.Get("frequency").(string)
	journeyViewSchedule := &platformclientv2.Journeyviewschedule{
		Frequency: &frequency,
	}
	return journeyViewSchedule
}
