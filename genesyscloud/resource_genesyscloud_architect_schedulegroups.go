package genesyscloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v56/platformclientv2"
)

func getAllArchitectScheduleGroups(_ context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(ResourceIDMetaMap)
	archAPI := platformclientv2.NewArchitectApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		scheduleGroups, _, getErr := archAPI.GetArchitectSchedulegroups(pageNum, pageSize, "", "", "", "", nil)
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of schedule groups: %v", getErr)
		}

		if scheduleGroups.Entities == nil || len(*scheduleGroups.Entities) == 0 {
			break
		}

		for _, scheduleGroup := range *scheduleGroups.Entities {
			resources[*scheduleGroup.Id] = &ResourceMeta{Name: *scheduleGroup.Name}
		}
	}

	return resources, nil
}

func architectScheduleGroupsExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllArchitectScheduleGroups),
		RefAttrs:         map[string]*RefAttrSettings{}, // No references
	}
}

func resourceArchitectScheduleGroups() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Architect Schedule Groups",

		CreateContext: createWithPooledClient(createArchitectScheduleGroups),
		ReadContext:   readWithPooledClient(readArchitectScheduleGroups),
		UpdateContext: updateWithPooledClient(updateArchitectScheduleGroups),
		DeleteContext: deleteWithPooledClient(deleteArchitectScheduleGroups),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the schedule group.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "Description of the schedule group.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"time_zone": {
				Description: "The timezone the schedules are a part of.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"open_schedules_id": {
				Description: "The schedules defining the hours an organization is open.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"closed_schedules_id": {
				Description: "The schedules defining the hours an organization is closed.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"holiday_schedules_id": {
				Description: "The schedules defining the hours an organization is closed for the holidays.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func createArchitectScheduleGroups(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	timeZone := d.Get("time_zone").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
	archAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	schedGroup := platformclientv2.Schedulegroup{
		Name:             &name,
		OpenSchedules:    buildSdkDomainEntityRefArr(d, "open_schedules_id"),
		ClosedSchedules:  buildSdkDomainEntityRefArr(d, "closed_schedules_id"),
		HolidaySchedules: buildSdkDomainEntityRefArr(d, "holiday_schedules_id"),
	}

	// Optional attributes
	if description != "" {
		schedGroup.Description = &description
	}

	if timeZone != "" {
		schedGroup.TimeZone = &timeZone
	}

	log.Printf("Creating schedule group %s", name)
	scheduleGroup, _, getErr := archAPI.PostArchitectSchedulegroups(schedGroup)
	if getErr != nil {
		return diag.Errorf("Failed to create schedule group %s | ERROR: %s", *scheduleGroup.Name, getErr)
	}

	d.SetId(*scheduleGroup.Id)

	log.Printf("Created schedule group %s %s", name, *scheduleGroup.Id)
	return readArchitectScheduleGroups(ctx, d, meta)
}

func readArchitectScheduleGroups(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	archAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	log.Printf("Reading schedule group %s", d.Id())

	return withRetriesForRead(ctx, 30*time.Second, d, func() *resource.RetryError {
		scheduleGroup, resp, getErr := archAPI.GetArchitectSchedulegroup(d.Id())
		if getErr != nil {
			if isStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to read schedule group %s: %s", d.Id(), getErr))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to read schedule group %s: %s", d.Id(), getErr))
		}

		d.Set("name", *scheduleGroup.Name)
		d.Set("description", nil)
		if scheduleGroup.Description != nil {
			d.Set("description", *scheduleGroup.Description)
		}

		if scheduleGroup.OpenSchedules != nil {
			d.Set("open_schedules_id", sdkDomainEntityRefArrToSet(*scheduleGroup.OpenSchedules))
		} else {
			d.Set("open_schedules_id", nil)
		}

		if scheduleGroup.ClosedSchedules != nil {
			d.Set("closed_schedules_id", sdkDomainEntityRefArrToSet(*scheduleGroup.ClosedSchedules))
		} else {
			d.Set("closed_schedules_id", nil)
		}

		if scheduleGroup.HolidaySchedules != nil {
			d.Set("holiday_schedules_id", sdkDomainEntityRefArrToSet(*scheduleGroup.HolidaySchedules))
		} else {
			d.Set("holiday_schedules_id", nil)
		}

		log.Printf("Read schedule group %s %s", d.Id(), *scheduleGroup.Name)
		return nil
	})
}

func updateArchitectScheduleGroups(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	timeZone := d.Get("time_zone").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
	archAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	diagErr := retryWhen(isVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current schedule group version
		scheduleGroup, resp, getErr := archAPI.GetArchitectSchedulegroup(d.Id())
		if getErr != nil {
			return resp, diag.Errorf("Failed to read schedule group %s: %s", d.Id(), getErr)
		}

		log.Printf("Updating schedule group %s", name)
		_, resp, putErr := archAPI.PutArchitectSchedulegroup(d.Id(), platformclientv2.Schedulegroup{
			Name:             &name,
			Version:          scheduleGroup.Version,
			Description:      &description,
			TimeZone:         &timeZone,
			OpenSchedules:    buildSdkDomainEntityRefArr(d, "open_schedules_id"),
			ClosedSchedules:  buildSdkDomainEntityRefArr(d, "closed_schedules_id"),
			HolidaySchedules: buildSdkDomainEntityRefArr(d, "holiday_schedules_id"),
		})
		if putErr != nil {
			return resp, diag.Errorf("Failed to update schedule group %s: %s", d.Id(), putErr)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Finished updating schedule group %s", name)
	time.Sleep(5 * time.Second)
	return readArchitectScheduleGroups(ctx, d, meta)
}

func deleteArchitectScheduleGroups(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	archAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	log.Printf("Deleting schedule %s", d.Id())
	_, err := archAPI.DeleteArchitectSchedulegroup(d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete schedule group %s: %s", d.Id(), err)
	}

	return withRetries(ctx, 30*time.Second, func() *resource.RetryError {
		scheduleGroup, resp, err := archAPI.GetArchitectSchedulegroup(d.Id())
		if err != nil {
			if isStatus404(resp) {
				// schedule group deleted
				log.Printf("Deleted schedule group %s", d.Id())
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("Error deleting schedule group %s: %s", d.Id(), err))
		}

		if *scheduleGroup.State == "deleted" {
			// schedule group deleted
			log.Printf("Deleted schedule group %s", d.Id())
			return nil
		}

		return resource.RetryableError(fmt.Errorf("Schedule group %s still exists", d.Id()))
	})
}
