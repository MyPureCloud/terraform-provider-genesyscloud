package architect_schedulegroups

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v121/platformclientv2"
	"log"
	"strings"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"time"
)

/*
The resource_genesyscloud_architect_schedulegroups.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthArchitectSchedulegroups retrieves all of the architect schedulegroups via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthArchitectSchedulegroupss(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	return nil, nil
}

// createArchitectSchedulegroups is used by the architect_schedulegroups resource to create Genesys cloud architect schedulegroups
func createArchitectSchedulegroups(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getArchitectSchedulegroupsProxy(sdkConfig)

	schedGroup := getArchitectScheduleGroupFromResourceData(d)

	log.Printf("Creating schedule group %s", *schedGroup.Name)
	scheduleGroup, err := proxy.createArchitectSchedulegroups(ctx, &schedGroup)
	if err != nil {
		msg := ""
		if strings.Contains(fmt.Sprintf("%v", err), "routing:schedule:add") {
			msg = "\nYou must have all divisions and future divisions selected in your OAuth client role"
		}
		return diag.Errorf("Failed to create schedule group %s | ERROR: %s%s", *schedGroup.Name, err, msg)
	}

	d.SetId(*scheduleGroup.Id)

	log.Printf("Created schedule group %s %s", *schedGroup.Name, *scheduleGroup.Id)
	return readArchitectSchedulegroups(ctx, d, meta)
}

// readArchitectSchedulegroups is used by the architect_schedulegroups resource to read an architect schedulegroups from genesys cloud
func readArchitectSchedulegroups(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getArchitectSchedulegroupsProxy(sdkConfig)

	log.Printf("Reading schedule group %s", d.Id())

	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		scheduleGroup, resp, getErr := proxy.getArchitectSchedulegroupsById(ctx, d.Id())
		if getErr != nil {
			if gcloud.IsStatus404ByInt(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read schedule group %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read schedule group %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceArchitectSchedulegroups())
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
			d.Set("open_schedules_id", gcloud.SdkDomainEntityRefArrToSet(*scheduleGroup.OpenSchedules))
		} else {
			d.Set("open_schedules_id", nil)
		}

		if scheduleGroup.ClosedSchedules != nil {
			d.Set("closed_schedules_id", gcloud.SdkDomainEntityRefArrToSet(*scheduleGroup.ClosedSchedules))
		} else {
			d.Set("closed_schedules_id", nil)
		}

		if scheduleGroup.HolidaySchedules != nil {
			d.Set("holiday_schedules_id", gcloud.SdkDomainEntityRefArrToSet(*scheduleGroup.HolidaySchedules))
		} else {
			d.Set("holiday_schedules_id", nil)
		}

		log.Printf("Read schedule group %s %s", d.Id(), *scheduleGroup.Name)
		return cc.CheckState()
	})
}

// updateArchitectSchedulegroups is used by the architect_schedulegroups resource to update an architect schedulegroups in Genesys Cloud
func updateArchitectSchedulegroups(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getArchitectSchedulegroupsProxy(sdkConfig)

	scheduleGroup := getArchitectScheduleGroupFromResourceData(d)

	log.Printf("Updating schedule group %s %s", *scheduleGroup.Name, d.Id())
	_, err := proxy.updateArchitectSchedulegroups(ctx, d.Id(), &scheduleGroup)

	if err != nil {
		msg := ""
		if strings.Contains(fmt.Sprintf("%v", err), "routing:schedule:add") {
			msg = "\nYou must have all divisions and future divisions selected in your OAuth client role"
		}
		return diag.Errorf("Failed to update schedule group %s: %s%s", d.Id(), err, msg)
	}

	log.Printf("Updated schedule group %s %s", *scheduleGroup.Name, d.Id())
	return readArchitectSchedulegroups(ctx, d, meta)
}

// deleteArchitectSchedulegroups is used by the architect_schedulegroups resource to delete an architect schedulegroups from Genesys cloud
func deleteArchitectSchedulegroups(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getArchitectSchedulegroupsProxy(sdkConfig)

	// DEVTOOLING-313: a schedule group linked to an IVR will not be able to be deleted until that IVR is deleted. Retrying here to make sure it is cleared properly.
	log.Printf("Deleting schedule group %s", d.Id())
	diagErr := gcloud.RetryWhen(gcloud.IsStatus409, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting schedule group %s", d.Id())
		resp, err := proxy.deleteArchitectSchedulegroups(ctx, d.Id())
		if err != nil {
			return resp, diag.Errorf("Failed to delete schedule group %s: %s", d.Id(), err)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return gcloud.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		scheduleGroup, resp, err := proxy.getArchitectSchedulegroupsById(ctx, d.Id())
		if err != nil {
			if gcloud.IsStatus404ByInt(resp) {
				// schedule group deleted
				log.Printf("Deleted schedule group %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting schedule group %s: %s", d.Id(), err))
		}

		if scheduleGroup.State != nil && *scheduleGroup.State == "deleted" {
			// schedule group deleted
			log.Printf("Deleted schedule group %s", d.Id())
			return nil
		}

		return retry.RetryableError(fmt.Errorf("Schedule group %s still exists", d.Id()))
	})
}

func getArchitectScheduleGroupFromResourceData(d *schema.ResourceData) platformclientv2.Schedulegroup {
	scheduleGroup := platformclientv2.Schedulegroup{
		Name:             platformclientv2.String(d.Get("name").(string)),
		OpenSchedules:    gcloud.BuildSdkDomainEntityRefArr(d, "open_schedules_id"),
		ClosedSchedules:  gcloud.BuildSdkDomainEntityRefArr(d, "closed_schedules_id"),
		HolidaySchedules: gcloud.BuildSdkDomainEntityRefArr(d, "holiday_schedules_id"),
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
