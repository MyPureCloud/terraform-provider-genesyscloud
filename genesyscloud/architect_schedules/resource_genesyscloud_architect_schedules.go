package architect_schedules

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/leekchan/timeutil"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"

	"github.com/mypurecloud/platform-client-sdk-go/v129/platformclientv2"
)

func getAllArchitectSchedules(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getArchitectSchedulesProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	schedules, proxyResponse, getErr := proxy.getAllArchitectSchedules(ctx)
	if getErr != nil {
		return nil, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to get page of schedule error: %s", getErr), proxyResponse)
	}

	for _, schedule := range *schedules {
		resources[*schedule.Id] = &resourceExporter.ResourceMeta{Name: *schedule.Name}
	}

	return resources, nil
}

// createArchitectSchedulegs is used by the architect_schedulegs resource to create Genesys cloud architect schedules
func createArchitectSchedules(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getArchitectSchedulesProxy(sdkConfig)

	name := d.Get("name").(string)
	divisionID := d.Get("division_id").(string)
	description := d.Get("description").(string)
	start := d.Get("start").(string)
	end := d.Get("end").(string)
	rrule := d.Get("rrule").(string)

	schedStart, err := time.Parse("2006-01-02T15:04:05.000000", start)
	if err != nil {
		return diag.Errorf("Failed to parse date %s: %s", start, err)
	}

	schedEnd, err := time.Parse("2006-01-02T15:04:05.000000", end)
	if err != nil {
		return diag.Errorf("Failed to parse date %s: %s", end, err)
	}

	schedule := platformclientv2.Schedule{
		Name:  &name,
		Start: &schedStart,
		End:   &schedEnd,
		Rrule: &rrule,
	}

	// Optional attributes
	if description != "" {
		schedule.Description = &description
	}

	if divisionID != "" {
		schedule.Division = &platformclientv2.Writabledivision{Id: &divisionID}
	}

	log.Printf("Creating schedule %s", *schedule.Name)
	scheduleResponse, proxyResponse, err := proxy.createArchitectSchedules(ctx, &schedule)
	if err != nil {
		msg := ""
		if strings.Contains(fmt.Sprintf("%v", err), "routing:schedule:add") {
			msg = "\nYou must have all divisions and future divisions selected in your OAuth client role"
		}

		return util.BuildAPIDiagnosticError("genesyscloud_archiect_schedules", fmt.Sprintf("Failed to create schedule %s | Error: %s MSG: %s", *scheduleResponse.Name, err, msg), proxyResponse)
	}

	d.SetId(*scheduleResponse.Id)

	log.Printf("Created schedule %s %s", name, *scheduleResponse.Id)
	return readArchitectSchedules(ctx, d, meta)
}

func readArchitectSchedules(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getArchitectSchedulesProxy(sdkConfig)

	log.Printf("Reading schedule %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		scheduleResponse, proxyResponse, err := proxy.getArchitectSchedulesById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(proxyResponse) {
				return retry.RetryableError(fmt.Errorf("failed to read schedule %s: %s", d.Id(), err))
			}
			return retry.NonRetryableError(fmt.Errorf("failed to read schedule %s: %s", d.Id(), err))
		}

		Start := new(string)
		if scheduleResponse.Start != nil {
			*Start = timeutil.Strftime(scheduleResponse.Start, "%Y-%m-%dT%H:%M:%S.%f")

		} else {
			Start = nil
		}

		End := new(string)
		if scheduleResponse.End != nil {
			*End = timeutil.Strftime(scheduleResponse.End, "%Y-%m-%dT%H:%M:%S.%f")

		} else {
			End = nil
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceArchitectSchedules(), constants.DefaultConsistencyChecks, resourceName)
		d.Set("name", *scheduleResponse.Name)
		d.Set("division_id", *scheduleResponse.Division.Id)
		d.Set("description", nil)
		if scheduleResponse.Description != nil {
			d.Set("description", *scheduleResponse.Description)
		}
		d.Set("start", Start)
		d.Set("end", End)
		d.Set("rrule", nil)
		if scheduleResponse.Rrule != nil {
			d.Set("rrule", *scheduleResponse.Rrule)
		}

		log.Printf("Read schedule %s %s", d.Id(), *scheduleResponse.Name)
		return cc.CheckState(d)
	})
}

func updateArchitectSchedules(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getArchitectSchedulesProxy(sdkConfig)

	name := d.Get("name").(string)
	divisionID := d.Get("division_id").(string)
	description := d.Get("description").(string)
	start := d.Get("start").(string)
	end := d.Get("end").(string)
	rrule := d.Get("rrule").(string)

	schedStart, err := time.Parse("2006-01-02T15:04:05.000000", start)
	if err != nil {
		return diag.Errorf("Failed to parse date %s: %s", start, err)
	}

	schedEnd, err := time.Parse("2006-01-02T15:04:05.000000", end)
	if err != nil {
		return diag.Errorf("Failed to parse date %s: %s", end, err)
	}

	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current schedule version
		scheduleResponse, proxyResponse, err := proxy.getArchitectSchedulesById(ctx, d.Id())

		if err != nil {
			return proxyResponse, util.BuildAPIDiagnosticError("genesyscloud_archiect_schedules", fmt.Sprintf("Failed to read schedule %s error: %s", d.Id(), err), proxyResponse)
		}

		log.Printf("Updating schedule %s", name)
		_, proxyUpdResponse, putErr := proxy.updateArchitectSchedules(ctx, d.Id(), &platformclientv2.Schedule{
			Name:        &name,
			Version:     scheduleResponse.Version,
			Division:    &platformclientv2.Writabledivision{Id: &divisionID},
			Description: &description,
			Start:       &schedStart,
			End:         &schedEnd,
			Rrule:       &rrule,
		})
		if putErr != nil {
			msg := ""
			if strings.Contains(fmt.Sprintf("%v", err), "routing:schedule:add") {
				msg = "\nYou must have all divisions and future divisions selected in your OAuth client role"
			}

			return proxyUpdResponse, util.BuildAPIDiagnosticError("genesyscloud_archiect_schedules", fmt.Sprintf("Failed to update schedule %s | Error: %s MSG: %s", *scheduleResponse.Name, putErr, msg), proxyUpdResponse)
		}
		return proxyUpdResponse, nil
	})

	if diagErr != nil {
		return diagErr
	}

	log.Printf("Finished updating schedule %s", name)
	return readArchitectSchedules(ctx, d, meta)
}

func deleteArchitectSchedules(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getArchitectSchedulesProxy(sdkConfig)

	// DEVTOOLING-311: a schedule linked to a schedule group will not be able to be deleted until that schedule group is deleted. Retryig here to make sure it is cleared properly.
	log.Printf("Deleting schedule %s", d.Id())
	diagErr := util.RetryWhen(util.IsStatus409, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting schedule %s", d.Id())
		proxyDelResponse, err := proxy.deleteArchitectSchedules(ctx, d.Id())
		if err != nil {
			return proxyDelResponse, util.BuildAPIDiagnosticError("genesyscloud_archiect_schedules", fmt.Sprintf("Failed to delete schedule %s error: %s", d.Id(), err), proxyDelResponse)
		}
		return proxyDelResponse, nil
	})

	if diagErr != nil {
		return diagErr
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		scheduleResponse, proxyGetResponse, err := proxy.getArchitectSchedulesById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(proxyGetResponse) {
				// schedule deleted
				log.Printf("Deleted schedule %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("error deleting schedule %s: %s", d.Id(), err))
		}

		if scheduleResponse.State != nil && *scheduleResponse.State == "deleted" {
			// schedule deleted
			log.Printf("Deleted group %s", d.Id())
			return nil
		}

		return retry.RetryableError(fmt.Errorf("schedule %s still exists", d.Id()))
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
