package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"terraform-provider-genesyscloud/genesyscloud/validators"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/leekchan/timeutil"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func getAllArchitectSchedules(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	archAPI := platformclientv2.NewArchitectApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		schedules, resp, getErr := archAPI.GetArchitectSchedules(pageNum, pageSize, "", "", "", nil)
		if getErr != nil {
			return nil, util.BuildAPIDiagnosticError("genesyscloud_architect_schedules", fmt.Sprintf("Failed to get page of schedules error: %s", getErr), resp)
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
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllArchitectSchedules),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"division_id": {RefType: "genesyscloud_auth_division"},
		},
		CustomValidateExports: map[string][]string{
			"rrule": {"rrule"},
		},
	}
}

func ResourceArchitectSchedules() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Architect Schedules",

		CreateContext: provider.CreateWithPooledClient(createArchitectSchedules),
		ReadContext:   provider.ReadWithPooledClient(readArchitectSchedules),
		UpdateContext: provider.UpdateWithPooledClient(updateArchitectSchedules),
		DeleteContext: provider.DeleteWithPooledClient(deleteArchitectSchedules),
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
				ValidateDiagFunc: validators.ValidateLocalDateTimes,
			},
			"end": {
				Description:      "Date time is represented as an ISO-8601 string without a timezone. For example: 2006-01-02T15:04:05.000000.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validators.ValidateLocalDateTimes,
			},
			"rrule": {
				Description:      "An iCal Recurrence Rule (RRULE) string. It is required to be set for schedules determining when upgrades to the Edge software can be applied.",
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validators.ValidateRrule,
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

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	archAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	schedStart, err := time.Parse("2006-01-02T15:04:05.000000", start)
	if err != nil {
		return util.BuildDiagnosticError("genesyscloud_architect_schedules", fmt.Sprintf("Failed to parse date %s", start), err)
	}

	schedEnd, err := time.Parse("2006-01-02T15:04:05.000000", end)
	if err != nil {
		return util.BuildDiagnosticError("genesyscloud_architect_schedules", fmt.Sprintf("Failed to parse date %s", end), err)
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
	schedule, resp, getErr := archAPI.PostArchitectSchedules(sched)
	if getErr != nil {
		msg := ""
		if strings.Contains(fmt.Sprintf("%v", getErr), "routing:schedule:add") {
			msg = "\nYou must have all divisions and future divisions selected in your OAuth client role"
		}

		return util.BuildAPIDiagnosticError("genesyscloud_archiect_schedules", fmt.Sprintf("Failed to create schedule %s | Error: %s MSG: %s", *sched.Name, err, msg), resp)
	}

	d.SetId(*schedule.Id)

	log.Printf("Created schedule %s %s", name, *schedule.Id)
	return readArchitectSchedules(ctx, d, meta)
}

func readArchitectSchedules(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	archAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceArchitectSchedules(), constants.DefaultConsistencyChecks, "genesyscloud_architect_schedules")

	log.Printf("Reading schedule %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		schedule, resp, getErr := archAPI.GetArchitectSchedule(d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_architect_schedules", fmt.Sprintf("Failed to read schedule %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_architect_schedules", fmt.Sprintf("Failed to read schedule %s | error: %s", d.Id(), getErr), resp))
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
		return cc.CheckState(d)
	})
}

func updateArchitectSchedules(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	divisionID := d.Get("division_id").(string)
	description := d.Get("description").(string)
	start := d.Get("start").(string)
	end := d.Get("end").(string)
	rrule := d.Get("rrule").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	archAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	schedStart, err := time.Parse("2006-01-02T15:04:05.000000", start)
	if err != nil {
		return util.BuildDiagnosticError("genesyscloud_architect_schedules", fmt.Sprintf("Failed to parse date %s", start), err)
	}

	schedEnd, err := time.Parse("2006-01-02T15:04:05.000000", end)
	if err != nil {
		return util.BuildDiagnosticError("genesyscloud_architect_schedules", fmt.Sprintf("Failed to parse date %s", end), err)
	}

	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current schedule version
		sched, resp, getErr := archAPI.GetArchitectSchedule(d.Id())

		if getErr != nil {
			return resp, util.BuildAPIDiagnosticError("genesyscloud_archiect_schedules", fmt.Sprintf("Failed to read schedule %s error: %s", d.Id(), err), resp)
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

			return resp, util.BuildAPIDiagnosticError("genesyscloud_archiect_schedules", fmt.Sprintf("Failed to update schedule %s | Error: %s MSG: %s", *sched.Name, putErr, msg), resp)
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
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	archAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	// DEVTOOLING-311: a schedule linked to a schedule group will not be able to be deleted until that schedule group is deleted. Retryig here to make sure it is cleared properly.
	log.Printf("Deleting schedule %s", d.Id())
	diagErr := util.RetryWhen(util.IsStatus409, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting schedule %s", d.Id())
		resp, err := archAPI.DeleteArchitectSchedule(d.Id())
		if err != nil {
			return resp, util.BuildAPIDiagnosticError("genesyscloud_archiect_schedules", fmt.Sprintf("Failed to delete schedule %s error: %s", d.Id(), err), resp)
		}
		return resp, nil
	})

	if diagErr != nil {
		return diagErr
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		schedule, resp, err := archAPI.GetArchitectSchedule(d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				// schedule deleted
				log.Printf("Deleted schedule %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_architect_schedules", fmt.Sprintf("Error deleting schedule %s | error: %s", d.Id(), err), resp))
		}

		if schedule.State != nil && *schedule.State == "deleted" {
			// schedule deleted
			log.Printf("Deleted group %s", d.Id())
			return nil
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_architect_schedules", fmt.Sprintf("Schedule %s still exists", d.Id()), resp))
	})
}

func GenerateArchitectSchedulesResource(
	schedResource1 string,
	name string,
	divisionId string,
	description string,
	start string,
	end string,
	rrule string) string {
	return fmt.Sprintf(`resource "genesyscloud_architect_schedules" "%s" {
		name = "%s"
		division_id = %s
		description = "%s"
		start = "%s"
		end = "%s"
		rrule = "%s"
	}
	`, schedResource1, name, divisionId, description, start, end, rrule)
}
