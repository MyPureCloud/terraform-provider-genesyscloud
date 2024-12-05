package architect_schedulegroups

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

/*
The resource_genesyscloud_architect_schedulegroups.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthArchitectSchedulegroups retrieves all of the architect schedulegroups via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthArchitectSchedulegroups(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getArchitectSchedulegroupsProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	scheduleGroups, proxyResponse, err := proxy.getAllArchitectSchedulegroups(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get page of schedule groups error: %s", err), proxyResponse)
	}

	for _, scheduleGroup := range *scheduleGroups {
		resources[*scheduleGroup.Id] = &resourceExporter.ResourceMeta{BlockLabel: *scheduleGroup.Name}
	}

	return resources, nil
}

// createArchitectSchedulegroups is used by the architect_schedulegroups resource to create Genesys cloud architect schedulegroups
func createArchitectSchedulegroups(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getArchitectSchedulegroupsProxy(sdkConfig)

	schedGroup := getArchitectScheduleGroupFromResourceData(d)

	log.Printf("Creating schedule group %s", *schedGroup.Name)
	scheduleGroup, proxyResponse, err := proxy.createArchitectSchedulegroups(ctx, &schedGroup)
	if err != nil {
		msg := ""
		if strings.Contains(fmt.Sprintf("%v", err), "routing:schedule:add") {
			msg = "\nYou must have all divisions and future divisions selected in your OAuth client role"
		}
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create schedule group %s | Error: %s %s", d.Id(), err, msg), proxyResponse)
	}

	d.SetId(*scheduleGroup.Id)

	log.Printf("Created schedule group %s %s", *schedGroup.Name, *scheduleGroup.Id)
	return readArchitectSchedulegroups(ctx, d, meta)
}

// readArchitectSchedulegroups is used by the architect_schedulegroups resource to read an architect schedulegroups from genesys cloud
func readArchitectSchedulegroups(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getArchitectSchedulegroupsProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceArchitectSchedulegroups(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading schedule group %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		scheduleGroup, proxyResponse, getErr := proxy.getArchitectSchedulegroupsById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(proxyResponse) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read schedule group %s | error: %s", d.Id(), getErr), proxyResponse))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read schedule group %s | error: %s", d.Id(), getErr), proxyResponse))
		}

		d.Set("name", *scheduleGroup.Name)
		d.Set("division_id", *scheduleGroup.Division.Id)
		d.Set("description", nil)
		if scheduleGroup.Description != nil {
			d.Set("description", *scheduleGroup.Description)
		}

		d.Set("time_zone", nil)
		if scheduleGroup.TimeZone != nil {
			d.Set("time_zone", *scheduleGroup.TimeZone)
		}

		if scheduleGroup.OpenSchedules != nil {
			d.Set("open_schedules_id", util.SdkDomainEntityRefArrToSet(*scheduleGroup.OpenSchedules))
		} else {
			d.Set("open_schedules_id", nil)
		}

		if scheduleGroup.ClosedSchedules != nil {
			d.Set("closed_schedules_id", util.SdkDomainEntityRefArrToSet(*scheduleGroup.ClosedSchedules))
		} else {
			d.Set("closed_schedules_id", nil)
		}

		if scheduleGroup.HolidaySchedules != nil {
			d.Set("holiday_schedules_id", util.SdkDomainEntityRefArrToSet(*scheduleGroup.HolidaySchedules))
		} else {
			d.Set("holiday_schedules_id", nil)
		}

		log.Printf("Read schedule group %s %s", d.Id(), *scheduleGroup.Name)
		return cc.CheckState(d)
	})
}

// updateArchitectSchedulegroups is used by the architect_schedulegroups resource to update an architect schedulegroups in Genesys Cloud
func updateArchitectSchedulegroups(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getArchitectSchedulegroupsProxy(sdkConfig)

	scheduleGroup := getArchitectScheduleGroupFromResourceData(d)

	log.Printf("Updating schedule group %s %s", *scheduleGroup.Name, d.Id())
	_, proxyResponse, err := proxy.updateArchitectSchedulegroups(ctx, d.Id(), &scheduleGroup)

	if err != nil {
		msg := ""
		if strings.Contains(fmt.Sprintf("%v", err), "routing:schedule:add") {
			msg = "\nYou must have all divisions and future divisions selected in your OAuth client role"
		}
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update schedule group %s | Error: %s %s", d.Id(), err, msg), proxyResponse)
	}

	log.Printf("Updated schedule group %s %s", *scheduleGroup.Name, d.Id())
	return readArchitectSchedulegroups(ctx, d, meta)
}

// deleteArchitectSchedulegroups is used by the architect_schedulegroups resource to delete an architect schedulegroups from Genesys cloud
func deleteArchitectSchedulegroups(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getArchitectSchedulegroupsProxy(sdkConfig)

	// DEVTOOLING-313: a schedule group linked to an IVR will not be able to be deleted until that IVR is deleted. Retrying here to make sure it is cleared properly.
	log.Printf("Deleting schedule group %s", d.Id())
	diagErr := util.RetryWhen(util.IsStatus409, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting schedule group %s", d.Id())
		proxyResponse, err := proxy.deleteArchitectSchedulegroups(ctx, d.Id())
		if err != nil {
			return proxyResponse, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete schedule group %s error: %s", d.Id(), err), proxyResponse)
		}
		return proxyResponse, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		scheduleGroup, proxyResponse, err := proxy.getArchitectSchedulegroupsById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(proxyResponse) {
				// schedule group deleted
				log.Printf("Deleted schedule group %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting schedule group %s | error: %s", d.Id(), err), proxyResponse))
		}

		if scheduleGroup.State != nil && *scheduleGroup.State == "deleted" {
			// schedule group deleted
			log.Printf("Deleted schedule group %s", d.Id())
			return nil
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Schedule group %s still exists", d.Id()), proxyResponse))
	})
}

func getArchitectScheduleGroupFromResourceData(d *schema.ResourceData) platformclientv2.Schedulegroup {
	scheduleGroup := platformclientv2.Schedulegroup{
		Name:             platformclientv2.String(d.Get("name").(string)),
		OpenSchedules:    util.BuildSdkDomainEntityRefArr(d, "open_schedules_id"),
		ClosedSchedules:  util.BuildSdkDomainEntityRefArr(d, "closed_schedules_id"),
		HolidaySchedules: util.BuildSdkDomainEntityRefArr(d, "holiday_schedules_id"),
	}

	divisionID := d.Get("division_id").(string)
	description := d.Get("description").(string)
	timeZone := d.Get("time_zone").(string)

	// Optional attributes
	if divisionID != "" {
		scheduleGroup.Division = &platformclientv2.Writabledivision{Id: &divisionID}
	}

	if description != "" {
		scheduleGroup.Description = &description
	}

	if timeZone != "" {
		scheduleGroup.TimeZone = &timeZone
	}

	return scheduleGroup
}
