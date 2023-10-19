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
	"github.com/leekchan/timeutil"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func getAllArchitectSchedules(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	archAPI := platformclientv2.NewArchitectApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		schedules, _, getErr := archAPI.GetArchitectSchedules(pageNum, pageSize, "", "", "", nil)
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of schedules: %v", getErr)
		}

		if schedules.Entities == nil || len(*schedules.Entities) == 0 {
			break
		}

		for _, schedule := range *schedules.Entities {
			resources[*schedule.Id] = &resourceExporter.ResourceMeta{Name: *schedule.Name}
		}
	}

	return resources, nil
}

func ArchitectSchedulesExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: GetAllWithPooledClient(getAllArchitectSchedules),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"division_id": {RefType: "genesyscloud_auth_division"},
		},
	}
}

func ResourceArchitectSchedules() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Architect Schedules",

		CreateContext: CreateWithPooledClient(createArchitectSchedules),
		ReadContext:   ReadWithPooledClient(readArchitectSchedules),
		UpdateContext: UpdateWithPooledClient(updateArchitectSchedules),
		DeleteContext: DeleteWithPooledClient(deleteArchitectSchedules),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the schedule.",
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
				Description: "Description of the schedule.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"start": {
				Description:      "Date time is represented as an ISO-8601 string without a timezone. For example: 2006-01-02T15:04:05.000000.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateLocalDateTimes,
			},
			"end": {
				Description:      "Date time is represented as an ISO-8601 string without a timezone. For example: 2006-01-02T15:04:05.000000.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateLocalDateTimes,
			},
			"rrule": {
				Description: "An iCal Recurrence Rule (RRULE) string.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func createArchitectSchedules(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	divisionID := d.Get("division_id").(string)
	description := d.Get("description").(string)
	start := d.Get("start").(string)
	end := d.Get("end").(string)
	rrule := d.Get("rrule").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	archAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	schedStart, err := time.Parse("2006-01-02T15:04:05.000000", start)
	if err != nil {
		return diag.Errorf("Failed to parse date %s: %s", start, err)
	}

	schedEnd, err := time.Parse("2006-01-02T15:04:05.000000", end)
	if err != nil {
		return diag.Errorf("Failed to parse date %s: %s", end, err)
	}

	sched := platformclientv2.Schedule{
		Name:  &name,
		Start: &schedStart,
		End:   &schedEnd,
		Rrule: &rrule,
	}

	// Optional attributes
	if description != "" {
		sched.Description = &description
	}

	if divisionID != "" {
		sched.Division = &platformclientv2.Writabledivision{Id: &divisionID}
	}

	log.Printf("Creating schedule %s", name)
	schedule, _, getErr := archAPI.PostArchitectSchedules(sched)
	if getErr != nil {
		msg := ""
		if strings.Contains(fmt.Sprintf("%v", getErr), "routing:schedule:add") {
			msg = "\nYou must have all divisions and future divisions selected in your OAuth client role"
		}
		return diag.Errorf("Failed to create schedule %s | ERROR: %s%s", *sched.Name, getErr, msg)
	}

	d.SetId(*schedule.Id)

	log.Printf("Created schedule %s %s", name, *schedule.Id)
	return readArchitectSchedules(ctx, d, meta)
}

func readArchitectSchedules(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	archAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	log.Printf("Reading schedule %s", d.Id())

	return WithRetriesForRead(ctx, d, func() *retry.RetryError {
		schedule, resp, getErr := archAPI.GetArchitectSchedule(d.Id())
		if getErr != nil {
			if IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read schedule %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read schedule %s: %s", d.Id(), getErr))
		}

		Start := new(string)
		if schedule.Start != nil {
			*Start = timeutil.Strftime(schedule.Start, "%Y-%m-%dT%H:%M:%S.%f")

		} else {
			Start = nil
		}

		End := new(string)
		if schedule.End != nil {
			*End = timeutil.Strftime(schedule.End, "%Y-%m-%dT%H:%M:%S.%f")

		} else {
			End = nil
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceArchitectSchedules())
		d.Set("name", *schedule.Name)
		d.Set("division_id", *schedule.Division.Id)
		d.Set("description", nil)
		if schedule.Description != nil {
			d.Set("description", *schedule.Description)
		}
		d.Set("start", Start)
		d.Set("end", End)
		d.Set("rrule", nil)
		if schedule.Rrule != nil {
			d.Set("rrule", *schedule.Rrule)
		}

		log.Printf("Read schedule %s %s", d.Id(), *schedule.Name)
		return cc.CheckState()
	})
}

func updateArchitectSchedules(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	divisionID := d.Get("division_id").(string)
	description := d.Get("description").(string)
	start := d.Get("start").(string)
	end := d.Get("end").(string)
	rrule := d.Get("rrule").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	archAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	schedStart, err := time.Parse("2006-01-02T15:04:05.000000", start)
	if err != nil {
		return diag.Errorf("Failed to parse date %s: %s", start, err)
	}

	schedEnd, err := time.Parse("2006-01-02T15:04:05.000000", end)
	if err != nil {
		return diag.Errorf("Failed to parse date %s: %s", end, err)
	}

	diagErr := RetryWhen(IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current schedule version
		sched, resp, getErr := archAPI.GetArchitectSchedule(d.Id())
		if getErr != nil {
			return resp, diag.Errorf("Failed to read schedule %s: %s", d.Id(), getErr)
		}

		log.Printf("Updating schedule %s", name)
		_, resp, putErr := archAPI.PutArchitectSchedule(d.Id(), platformclientv2.Schedule{
			Name:        &name,
			Version:     sched.Version,
			Division:    &platformclientv2.Writabledivision{Id: &divisionID},
			Description: &description,
			Start:       &schedStart,
			End:         &schedEnd,
			Rrule:       &rrule,
		})
		if putErr != nil {
			msg := ""
			if strings.Contains(fmt.Sprintf("%v", getErr), "routing:schedule:add") {
				msg = "\nYou must have all divisions and future divisions selected in your OAuth client role"
			}
			return resp, diag.Errorf("Failed to update schedule %s | ERROR: %s%s", *sched.Name, getErr, msg)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Finished updating schedule %s", name)
	return readArchitectSchedules(ctx, d, meta)
}

func deleteArchitectSchedules(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	archAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	log.Printf("Deleting schedule %s", d.Id())
	_, err := archAPI.DeleteArchitectSchedule(d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete schedule %s: %s", d.Id(), err)
	}

	return WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		schedule, resp, err := archAPI.GetArchitectSchedule(d.Id())
		if err != nil {
			if IsStatus404(resp) {
				// schedule deleted
				log.Printf("Deleted schedule %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting schedule %s: %s", d.Id(), err))
		}

		if schedule.State != nil && *schedule.State == "deleted" {
			// schedule deleted
			log.Printf("Deleted group %s", d.Id())
			return nil
		}

		return retry.RetryableError(fmt.Errorf("Schedule %s still exists", d.Id()))
	})
}
