package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func getAllArchitectScheduleGroups(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
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
			resources[*scheduleGroup.Id] = &resourceExporter.ResourceMeta{Name: *scheduleGroup.Name}
		}
	}

	return resources, nil
}

func ArchitectScheduleGroupsExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: GetAllWithPooledClient(getAllArchitectScheduleGroups),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"division_id":          {RefType: "genesyscloud_auth_division"},
			"open_schedules_id":    {RefType: "genesyscloud_architect_schedules"},
			"closed_schedules_id":  {RefType: "genesyscloud_architect_schedules"},
			"holiday_schedules_id": {RefType: "genesyscloud_architect_schedules"},
		},
	}
}

func ResourceArchitectScheduleGroups() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Architect Schedule Groups",

		CreateContext: CreateWithPooledClient(createArchitectScheduleGroups),
		ReadContext:   ReadWithPooledClient(readArchitectScheduleGroups),
		UpdateContext: UpdateWithPooledClient(updateArchitectScheduleGroups),
		DeleteContext: DeleteWithPooledClient(deleteArchitectScheduleGroups),
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
			"division_id": {
				Description: "The division to which this schedule group will belong. If not set, the home division will be used. If set, you must have all divisions and future divisions selected in your OAuth client role",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
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
				Required:    true,
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
	divisionID := d.Get("division_id").(string)
	description := d.Get("description").(string)
	timeZone := d.Get("time_zone").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	archAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	schedGroup := platformclientv2.Schedulegroup{
		Name:             &name,
		OpenSchedules:    BuildSdkDomainEntityRefArr(d, "open_schedules_id"),
		ClosedSchedules:  BuildSdkDomainEntityRefArr(d, "closed_schedules_id"),
		HolidaySchedules: BuildSdkDomainEntityRefArr(d, "holiday_schedules_id"),
	}

	// Optional attributes
	if divisionID != "" {
		schedGroup.Division = &platformclientv2.Writabledivision{Id: &divisionID}
	}

	if description != "" {
		schedGroup.Description = &description
	}

	if timeZone != "" {
		schedGroup.TimeZone = &timeZone
	}

	log.Printf("Creating schedule group %s", name)
	scheduleGroup, _, getErr := archAPI.PostArchitectSchedulegroups(schedGroup)
	if getErr != nil {
		msg := ""
		if strings.Contains(fmt.Sprintf("%v", getErr), "routing:schedule:add") {
			msg = "\nYou must have all divisions and future divisions selected in your OAuth client role"
		}
		return diag.Errorf("Failed to create schedule group %s | ERROR: %s%s", *schedGroup.Name, getErr, msg)
	}

	d.SetId(*scheduleGroup.Id)

	log.Printf("Created schedule group %s %s", name, *scheduleGroup.Id)
	return readArchitectScheduleGroups(ctx, d, meta)
}

func readArchitectScheduleGroups(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	archAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	log.Printf("Reading schedule group %s", d.Id())

	return WithRetriesForRead(ctx, d, func() *retry.RetryError {
		scheduleGroup, resp, getErr := archAPI.GetArchitectSchedulegroup(d.Id())
		if getErr != nil {
			if IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read schedule group %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read schedule group %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceArchitectScheduleGroups())
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
		return cc.CheckState()
	})
}

func updateArchitectScheduleGroups(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	divisionID := d.Get("division_id").(string)
	description := d.Get("description").(string)
	timeZone := d.Get("time_zone").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	archAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	diagErr := RetryWhen(IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current schedule group version
		scheduleGroup, resp, getErr := archAPI.GetArchitectSchedulegroup(d.Id())
		if getErr != nil {
			return resp, diag.Errorf("Failed to read schedule group %s: %s", d.Id(), getErr)
		}

		log.Printf("Updating schedule group %s", name)
		_, resp, putErr := archAPI.PutArchitectSchedulegroup(d.Id(), platformclientv2.Schedulegroup{
			Name:             &name,
			Division:         &platformclientv2.Writabledivision{Id: &divisionID},
			Version:          scheduleGroup.Version,
			Description:      &description,
			TimeZone:         &timeZone,
			OpenSchedules:    BuildSdkDomainEntityRefArr(d, "open_schedules_id"),
			ClosedSchedules:  BuildSdkDomainEntityRefArr(d, "closed_schedules_id"),
			HolidaySchedules: BuildSdkDomainEntityRefArr(d, "holiday_schedules_id"),
		})
		if putErr != nil {
			msg := ""
			if strings.Contains(fmt.Sprintf("%v", getErr), "routing:schedule:add") {
				msg = "\nYou must have all divisions and future divisions selected in your OAuth client role"
			}
			return resp, diag.Errorf("Failed to update schedule group %s: %s%s", d.Id(), putErr, msg)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Finished updating schedule group %s", name)
	return readArchitectScheduleGroups(ctx, d, meta)
}

func deleteArchitectScheduleGroups(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	archAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	log.Printf("Deleting schedule %s", d.Id())
	_, err := archAPI.DeleteArchitectSchedulegroup(d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete schedule group %s: %s", d.Id(), err)
	}

	return WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		scheduleGroup, resp, err := archAPI.GetArchitectSchedulegroup(d.Id())
		if err != nil {
			if IsStatus404(resp) {
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
