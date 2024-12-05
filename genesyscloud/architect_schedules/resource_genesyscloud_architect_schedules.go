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
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

const timeFormat = "2006-01-02T15:04:05.000000"

func getAllArchitectSchedules(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getArchitectSchedulesProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	schedules, proxyResponse, getErr := proxy.getAllArchitectSchedules(ctx)
	if getErr != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get page of schedule error: %s", getErr), proxyResponse)
	}

	for _, schedule := range *schedules {
		resources[*schedule.Id] = &resourceExporter.ResourceMeta{BlockLabel: *schedule.Name}
	}

	return resources, nil
}

// createArchitectSchedules is used by the architect_schedules resource to create Genesys cloud architect schedules
func createArchitectSchedules(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getArchitectSchedulesProxy(sdkConfig)

	name := d.Get("name").(string)
	divisionID := d.Get("division_id").(string)
	description := d.Get("description").(string)
	start := d.Get("start").(string)
	end := d.Get("end").(string)
	rrule := d.Get("rrule").(string)

	//The first parameter of the Parse() method specifies the date and time format/layout that should be used to interpret the second parameter.
	schedStart, err := time.Parse(timeFormat, start)
	if err != nil {
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Failed to parse date %s", start), err)
	}

	schedEnd, err := time.Parse(timeFormat, end)
	if err != nil {
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Failed to parse date %s", end), err)
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

		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create schedule %s | Error: %s. %s", name, err, msg), proxyResponse)
	}

	d.SetId(*scheduleResponse.Id)

	log.Printf("Created schedule %s %s", name, *scheduleResponse.Id)
	return readArchitectSchedules(ctx, d, meta)
}

func readArchitectSchedules(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getArchitectSchedulesProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceArchitectSchedules(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading schedule %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		scheduleResponse, proxyResponse, err := proxy.getArchitectSchedulesById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(proxyResponse) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read schedule %s | error: %s", d.Id(), err), proxyResponse))
			}
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read schedule %s | error: %s", d.Id(), err), proxyResponse))
		}

		start := new(string)
		if scheduleResponse.Start != nil {
			*start = timeutil.Strftime(scheduleResponse.Start, "%Y-%m-%dT%H:%M:%S.%f")

		} else {
			start = nil
		}

		end := new(string)
		if scheduleResponse.End != nil {
			*end = timeutil.Strftime(scheduleResponse.End, "%Y-%m-%dT%H:%M:%S.%f")

		} else {
			end = nil
		}

		resourcedata.SetNillableValue(d, "name", scheduleResponse.Name)
		resourcedata.SetNillableValue(d, "division_id", scheduleResponse.Division.Id)
		resourcedata.SetNillableValue(d, "description", scheduleResponse.Description)
		resourcedata.SetNillableValue(d, "start", start)
		resourcedata.SetNillableValue(d, "end", end)
		resourcedata.SetNillableValue(d, "rrule", scheduleResponse.Rrule)

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

	//The first parameter of the Parse() method specifies the date and time format/layout that should be used to interpret the second parameter.
	schedStart, err := time.Parse(timeFormat, start)
	if err != nil {
		return diag.Errorf("Failed to parse date %s: %s", start, err)
	}

	schedEnd, err := time.Parse(timeFormat, end)
	if err != nil {
		return diag.Errorf("Failed to parse date %s: %s", end, err)
	}

	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current schedule version
		scheduleResponse, proxyResponse, err := proxy.getArchitectSchedulesById(ctx, d.Id())

		if err != nil {
			return proxyResponse, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read schedule %s error: %s", d.Id(), err), proxyResponse)
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

			return proxyUpdResponse, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update schedule %s | Error: %s. %s", name, putErr, msg), proxyUpdResponse)
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
			return proxyDelResponse, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete schedule %s error: %s", d.Id(), err), proxyDelResponse)
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
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting schedule %s | error: %s", d.Id(), err), proxyGetResponse))
		}

		if scheduleResponse.State != nil && *scheduleResponse.State == "deleted" {
			// schedule deleted
			log.Printf("Deleted group %s", d.Id())
			return nil
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Schedule %s still exists", d.Id()), proxyGetResponse))
	})
}

func GenerateArchitectSchedulesResource(
	schedResourceLabel string,
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
	`, schedResourceLabel, name, divisionId, description, start, end, rrule)
}
