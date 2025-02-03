package journey_view_schedule

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
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
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get journey view schedules: %v", err), resp)
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

	JourneyViewId := d.Get("journey_view_id").(string)
	journeyViewSchedule := makeJourneyViewScheduleFromSchema(d)

	log.Printf("Creating Schedule for jounery view id: %s", JourneyViewId)
	journeyViewSchedule, resp, err := gp.createJourneyViewSchedule(ctx, JourneyViewId, journeyViewSchedule)

	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create schedule for journey view id %s: %s", JourneyViewId, err), resp)
	}

	// The create API returns 201 with an empty response body. So the schedule id is nil
	// The journey view and its schedule is a 1-1 map
	// The schedule entity id is identical to the journey view id, using this to set resource Id
	if journeyViewSchedule.Id == nil {
		d.SetId(JourneyViewId)
	} else {
		d.SetId(*journeyViewSchedule.Id)
	}

	log.Printf("Created schedule for jounery view id: %s", JourneyViewId)
	return readJourneyViewSchedule(ctx, d, meta)
}

// readJourneyViewSchedule is used by the journey_view_schedule resource to read an journey view schedule from genesys cloud
func readJourneyViewSchedule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	JourneyViewId := d.Get("journey_view_id").(string)
	if JourneyViewId == "" {
		JourneyViewId = d.Id()
	}

	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceJourneyViewSchedule(), constants.ConsistencyChecks(), ResourceType)
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	gp := getJourneyViewScheduleProxy(sdkConfig)

	log.Printf("Reading schedule for journey view id: %s", JourneyViewId)

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		journeyViewSchedule, resp, err := gp.getJourneyViewScheduleByViewId(ctx, JourneyViewId)
		if err != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to get journeyView with viewId %s | error: %s", JourneyViewId, err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to get journeyView with viewId %s | error: %s", JourneyViewId, err), resp))
		}

		resourcedata.SetNillableValue(d, "frequency", journeyViewSchedule.Frequency)
		resourcedata.SetNillableValue(d, "journey_view_id", journeyViewSchedule.Id)

		log.Printf("Read journey view schedule %s", d.Id())
		return cc.CheckState(d)
	})
}

// updateJourneyViewSchedule is used by the journey_view_schedule resource to update a journey view schedule in Genesys Cloud
func updateJourneyViewSchedule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getJourneyViewScheduleProxy(sdkConfig)

	JourneyViewId := d.Get("journey_view_id").(string)
	journeyViewSchedule := makeJourneyViewScheduleFromSchema(d)

	log.Printf("Updating schedule for journey view id %s ", JourneyViewId)

	journeyViewSchedule, resp, err := proxy.updateJourneyViewSchedule(ctx, JourneyViewId, journeyViewSchedule)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update schedule for journey view id %s:%s", JourneyViewId, err), resp)
	}

	log.Printf("Updated schedule for journey view id %s", JourneyViewId)
	return readJourneyViewSchedule(ctx, d, meta)
}

// deleteJourneyViewSchedule is used by the journey_view_schedule resource to delete a journey view schedule from Genesys cloud
func deleteJourneyViewSchedule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	JourneyViewId := d.Get("journey_view_id").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getJourneyViewScheduleProxy(sdkConfig)

	log.Printf("Deleting schedule for journey view id: %s", JourneyViewId)

	resp, err := proxy.deleteJourneyViewSchedule(ctx, JourneyViewId)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete schedule for journey view id %s: %s", JourneyViewId, err), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getJourneyViewScheduleByViewId(ctx, d.Id())

		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted schedule for journey view id %s", JourneyViewId)
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting schedule for journey view id %s: %s", JourneyViewId, err), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Schedule for journey view id %s still exists", JourneyViewId), resp))
	})
}

func makeJourneyViewScheduleFromSchema(d *schema.ResourceData) *platformclientv2.Journeyviewschedule {
	frequency := d.Get("frequency").(string)
	journeyViewSchedule := &platformclientv2.Journeyviewschedule{
		Frequency: &frequency,
	}
	return journeyViewSchedule
}
