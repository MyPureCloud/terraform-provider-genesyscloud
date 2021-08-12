package genesyscloud

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v48/platformclientv2"
)

func getAllArchitectSchedules(ctx context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(ResourceIDMetaMap)
	archAPI := platformclientv2.NewArchitectApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		schedules, _, getErr := archAPI.GetArchitectSchedules(pageNum, 100, "", "", "", nil)
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
			},
			"description": {
				Description: "Description of the schedule.",
				Type:        schema.TypeString,
				Optional:    true,
			},
            "start": {
				Description: "Date time is represented as an ISO-8601 string without a timezone.",
				Type:         schema.TypeString,
				Required:     true,
			},
            "end": {
				Description: "Date time is represented as an ISO-8601 string without a timezone.",
				Type:         schema.TypeString,
				Required:     true,
			},
            "rrule": {
				Description: "An iCal Recurrence Rule (RRULE) string.",
				Type:         schema.TypeString,
				Required:     true,
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

	schedStart, err := time.Parse("2006-01-02T15:04:05.000", start)
	if err != nil {
		return diag.Errorf("Failed to parse date %s: %s", start, err)
	}

	schedEnd, err := time.Parse("2006-01-02T15:04:05.000", end)
	if err != nil {
		return diag.Errorf("Failed to parse date %s: %s", end, err)
	}

	sched := platformclientv2.Schedule{
		Name:       		&name,
		Description:      	&description,
		Start:				&schedStart,
		End:				&schedEnd,
		Rrule:				&rrule,
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

	schedule, resp, getErr := archAPI.GetArchitectSchedule(d.Id())
	if getErr != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read schedule %s: %s", d.Id(), getErr)
	}

	d.Set("name", *schedule.Name)
	d.Set("description", *schedule.Description)
	d.Set("start", *schedule.Start)
	d.Set("end", *schedule.End)
	d.Set("rrule", *schedule.Rrule)

	log.Printf("Read schedule %s %s", d.Id(), *schedule.Name)
	return nil
}

func updateArchitectSchedules(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	start := d.Get("start").(string)
	end := d.Get("end").(string)
	rrule := d.Get("rrule").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
	archAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	schedStart, err := time.Parse("2006-01-02T15:04:05.000", start)
	if err != nil {
		return diag.Errorf("Failed to parse date %s: %s", start, err)
	}

	schedEnd, err := time.Parse("2006-01-02T15:04:05.000", end)
	if err != nil {
		return diag.Errorf("Failed to parse date %s: %s", end, err)
	}

	sched := platformclientv2.Schedule{
		Name:       		&name,
		Description:      	&description,
		Start:				&schedStart,
		End:				&schedEnd,
		Rrule:				&rrule,
	}

	log.Printf("Updating schedule %s", name)
	schedule, _, getErr := archAPI.PutArchitectSchedule(d.Id(), sched)
	if getErr != nil {
		return diag.Errorf("Failed to update schedule %s: %s", name, getErr)
	}

	d.SetId(*schedule.Id)

	log.Printf("Finished updating schedule %s %s", name, *schedule.Id)
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

	log.Printf("Deleted schedule %s", d.Id())
	return nil
}