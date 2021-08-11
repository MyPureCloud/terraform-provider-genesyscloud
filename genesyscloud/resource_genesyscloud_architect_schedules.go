package genesyscloud

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
			"version": {
				Description: "Schedule's current version.",
				Type:        schema.TypeInt,
				Optional:    true,
			},
            "date_created": {
				Description: "The date the schedule was created.",
				Type:         schema.TypeString,
                ValidateFunc: validation.IsRFC3339Time,
				Optional:     true,
			},
            "date_modified": {
				Description: "The date of last modification to the schedule.",
				Type:         schema.TypeString,
                ValidateFunc: validation.IsRFC3339Time,
				Optional:     true,
			},
            "modified_by": {
				Description: "ID of the user that last modified the schedule.",
				Type:        schema.TypeString,
				Optional:    true,
			},
            "created_by": {
				Description: "ID of the user that created the schedule.",
				Type:        schema.TypeString,
				Optional:    true,
			},
            "state": {
				Description: "Indicates if the schedule is active, inactive or deleted.",
				Type:        schema.TypeString,
				Optional:    true,
			},
            "modified_by_app": {
				Description: "Application that last modified the schedule.",
				Type:        schema.TypeString,
				Optional:    true,
			},
            "created_by_app": {
				Description: "Application that created the schedule.",
				Type:        schema.TypeString,
				Optional:    true,
			},
            "start": {
				Description: "Date time is represented as an ISO-8601 string without a timezone.",
				Type:         schema.TypeString,
                ValidateFunc: validation.IsRFC3339Time,
				Required:     true,
			},
            "end": {
				Description: "Date time is represented as an ISO-8601 string without a timezone.",
				Type:         schema.TypeString,
                ValidateFunc: validation.IsRFC3339Time,
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
	version := d.Get("version").(int)
	dateCreated := d.Get("date_created").(string)
	dateModified := d.Get("date_modified").(string)
	modifiedBy := d.Get("modified_by").(string)
	createdBy := d.Get("created_by").(string)
	state := d.Get("state").(string)
	modifiedByApp := d.Get("modified_by_app").(string)
	createdByApp := d.Get("created_by_app").(string)
	start := d.Get("start").(string)
	end := d.Get("end").(string)
	rrule := d.Get("rrule").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
	archAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	test := time.Parse("2006-01-02T00:00:00.000Z", dateCreated)

	sched := platformclientv2.Schedule{
		Name:       		&name,
		Description:      	&description,
		Version:  			&version,
		DateCreated: 		test,
		DateModified:      	&dateModified,
		ModifiedBy:			&modifiedBy,
		CreatedBy:			&createdBy,
		State:				&state,
		ModifiedByApp:		&modifiedByApp,
		CreatedByApp:		&createdByApp,
		Start:				&start,
		End:				&end,
		Rrule:				&rrule,
	}

	log.Printf("Creating schedule %s", name)
	schedule, resp, getErr := archAPI.PostArchitectSchedules(sched)
	if getErr != nil {
		return diag.Errorf("Failed to create schedule %s: %s", name, getErr)
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
	d.Set("version", *schedule.Version)
	d.Set("date_created", *schedule.DateCreated)
	d.Set("date_modified", *schedule.DateModified)
	d.Set("modified_by", *schedule.ModifiedBy)
	d.Set("created_by", *schedule.CreatedBy)
	d.Set("state", *schedule.State)
	d.Set("modified_by_app", *schedule.ModifiedByApp)
	d.Set("created_by_app", *schedule.CreatedByApp)
	d.Set("start", *schedule.Start)
	d.Set("end", *schedule.End)
	d.Set("rrule", *schedule.Rrule)

	log.Printf("Read schedule %s %s", d.Id(), *schedule.Name)
	return nil
}

func updateArchitectSchedules(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	version := d.Get("version").(int)
	dateCreated := d.Get("date_created").(string)
	dateModified := d.Get("date_modified").(string)
	modifiedBy := d.Get("modified_by").(string)
	createdBy := d.Get("created_by").(string)
	state := d.Get("state").(string)
	modifiedByApp := d.Get("modified_by_app").(string)
	createdByApp := d.Get("created_by_app").(string)
	start := d.Get("start").(string)
	end := d.Get("end").(string)
	rrule := d.Get("rrule").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
	archAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	sched := platformclientv2.Schedule{
		Name:       		&name,
		Description:      	&description,
		Version:  			&version,
		DateCreated: 		&dateCreated,
		DateModified:      	&dateModified,
		ModifiedBy:			&modifiedBy,
		CreatedBy:			&createdBy,
		State:				&state,
		ModifiedByApp:		&modifiedByApp,
		CreatedByApp:		&createdByApp,
		Start:				&start,
		End:				&end,
		Rrule:				&rrule,
	}

	log.Printf("Updating schedule %s", name)
	schedule, resp, getErr := archAPI.PutArchitectSchedule(d.Id(), sched)
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