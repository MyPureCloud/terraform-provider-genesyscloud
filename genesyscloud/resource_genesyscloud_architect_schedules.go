package genesyscloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/leekchan/timeutil"
	"github.com/mypurecloud/platform-client-sdk-go/v56/platformclientv2"
)

func getAllArchitectSchedules(_ context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(ResourceIDMetaMap)
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
			resources[*schedule.Id] = &ResourceMeta{Name: *schedule.Name}
		}
	}

	return resources, nil
}

func architectSchedulesExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllArchitectSchedules),
		RefAttrs:         map[string]*RefAttrSettings{}, // No references
	}
}

func resourceArchitectSchedules() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Architect Schedules",

		CreateContext: createWithPooledClient(createArchitectSchedules),
		ReadContext:   readWithPooledClient(readArchitectSchedules),
		UpdateContext: updateWithPooledClient(updateArchitectSchedules),
		DeleteContext: deleteWithPooledClient(deleteArchitectSchedules),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the schedule.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
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
	description := d.Get("description").(string)
	start := d.Get("start").(string)
	end := d.Get("end").(string)
	rrule := d.Get("rrule").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
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

	log.Printf("Creating schedule %s", name)
	schedule, _, getErr := archAPI.PostArchitectSchedules(sched)
	if getErr != nil {
		return diag.Errorf("Failed to create schedule %s: | Start: %s, | End: %s, | ERROR: %s", *sched.Name, *sched.Start, *sched.End, getErr)
	}

	d.SetId(*schedule.Id)

	log.Printf("Created schedule %s %s", name, *schedule.Id)
	return readArchitectSchedules(ctx, d, meta)
}

func readArchitectSchedules(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	archAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	log.Printf("Reading schedule %s", d.Id())

	return withRetriesForRead(ctx, 30*time.Second, d, func() *resource.RetryError {
		schedule, resp, getErr := archAPI.GetArchitectSchedule(d.Id())
		if getErr != nil {
			if isStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to read schedule %s: %s", d.Id(), getErr))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to read schedule %s: %s", d.Id(), getErr))
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

		d.Set("name", *schedule.Name)
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
		return nil
	})
}

func updateArchitectSchedules(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	start := d.Get("start").(string)
	end := d.Get("end").(string)
	rrule := d.Get("rrule").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
	archAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	schedStart, err := time.Parse("2006-01-02T15:04:05.000000", start)
	if err != nil {
		return diag.Errorf("Failed to parse date %s: %s", start, err)
	}

	schedEnd, err := time.Parse("2006-01-02T15:04:05.000000", end)
	if err != nil {
		return diag.Errorf("Failed to parse date %s: %s", end, err)
	}

	diagErr := retryWhen(isVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current schedule version
		sched, resp, getErr := archAPI.GetArchitectSchedule(d.Id())
		if getErr != nil {
			return resp, diag.Errorf("Failed to read schedule %s: %s", d.Id(), getErr)
		}

		log.Printf("Updating schedule %s", name)
		_, resp, putErr := archAPI.PutArchitectSchedule(d.Id(), platformclientv2.Schedule{
			Name:        &name,
			Version:     sched.Version,
			Description: &description,
			Start:       &schedStart,
			End:         &schedEnd,
			Rrule:       &rrule,
		})
		if putErr != nil {
			return resp, diag.Errorf("Failed to update schedule %s: %s", d.Id(), putErr)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Finished updating schedule %s", name)
	time.Sleep(5 * time.Second)
	return readArchitectSchedules(ctx, d, meta)
}

func deleteArchitectSchedules(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	archAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	log.Printf("Deleting schedule %s", d.Id())
	_, err := archAPI.DeleteArchitectSchedule(d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete schedule %s: %s", d.Id(), err)
	}

	return withRetries(ctx, 30*time.Second, func() *resource.RetryError {
		schedule, resp, err := archAPI.GetArchitectSchedule(d.Id())
		if err != nil {
			if isStatus404(resp) {
				// schedule deleted
				log.Printf("Deleted schedule %s", d.Id())
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("Error deleting schedule %s: %s", d.Id(), err))
		}

		if schedule.State != nil && *schedule.State == "deleted" {
			// schedule deleted
			log.Printf("Deleted group %s", d.Id())
			return nil
		}

		return resource.RetryableError(fmt.Errorf("Schedule %s still exists", d.Id()))
	})
}
