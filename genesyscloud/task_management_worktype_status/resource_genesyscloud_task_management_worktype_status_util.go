package task_management_worktype_status

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/util"
)

// ModifyStatusIdStateValue will change the statusId before it is saved in the state file.
// The worktype_status resource saves the status id as <worktypeId>/<statusId>.
// We only want to save the statusId in the state as the api will only return the status id
// and this cause would 'plan not empty' if we save the id as <worktypeId>/<statusId>
func ModifyStatusIdStateValue(id interface{}) string {
	statusId := id.(string)
	if strings.Contains(statusId, "/") {
		return strings.Split(statusId, "/")[1]
	}

	return statusId
}

// SplitWorktypeStatusTerraformId will split the status resource id which is in the form
// <worktypeId>/<statusId> into just the worktypeId and statusId string
func SplitWorktypeStatusTerraformId(id string) (worktypeId string, statusId string) {
	return strings.Split(id, "/")[0], strings.Split(id, "/")[1]
}

// validateSchema checks if status_transition_delay_seconds was provided with default_destination_status_id
func validateSchema(d *schema.ResourceData) error {
	if d.Get("default_destination_status_id").(string) != "" {
		if d.Get("status_transition_delay_seconds").(int) == 0 {
			return fmt.Errorf("status_transition_delay_seconds is required with default_destination_status_id")
		}
	}

	return nil
}

func updateWorktypeDefaultStatus(ctx context.Context, proxy *taskManagementWorktypeStatusProxy, worktypeId string, statusId string) diag.Diagnostics {
	worktypeUpdate := platformclientv2.Worktypeupdate{
		DefaultStatusId: &statusId,
	}

	_, resp, err := proxy.worktypeProxy.UpdateTaskManagementWorktype(ctx, worktypeId, &worktypeUpdate)
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to update worktype %s with default status %s.", worktypeId, statusId), resp)
	}

	return nil
}

func GenerateWorktypeStatusResource(
	resourceName,
	workTypeId,
	name,
	category,
	description string,
	defaultDestinationStatusId string,
	statusTransitionTime string,
	attrs ...string,
) string {
	return fmt.Sprintf(
		`resource "genesyscloud_task_management_worktype_status" "%s" {
		worktype_id = %s
		name = "%s"
		category = "%s"
		description = "%s"
		default_destination_status_id = %s
		status_transition_time = "%s"
		%s
	}
`, resourceName, workTypeId, name, category, description, defaultDestinationStatusId, statusTransitionTime, strings.Join(attrs, "\n"))
}

// ValidateStatusIds will check that two status ids are the same
// We need this to handle situations where a reference to a status resource is used. In this case
// the id will be in the format <worktypeId>/<statusId> which is allowed but there terraform function cant check for this
func ValidateStatusIds(statusResource1 string, key1 string, statusResource2 string, key2 string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		status1, ok := state.RootModule().Resources[statusResource1]
		if !ok {
			return fmt.Errorf("failed to find status %s in state", statusResource1)
		}

		status2, ok := state.RootModule().Resources[statusResource2]
		if !ok {
			return fmt.Errorf("failed to find status %s in state", statusResource1)
		}

		status1KeyValue := status1.Primary.Attributes[key1]
		if strings.Contains(status1KeyValue, "/") {
			_, status1KeyValue = SplitWorktypeStatusTerraformId(status1KeyValue)
		}

		status2KeyValue := status2.Primary.Attributes[key2]
		if strings.Contains(status2KeyValue, "/") {
			_, status2KeyValue = SplitWorktypeStatusTerraformId(status2KeyValue)
		}

		if status1KeyValue != status2KeyValue {
			attr1 := statusResource1 + "." + key1
			attr2 := statusResource2 + "." + key2
			return fmt.Errorf("%s not equal to %s\n %s = %s\n %s = %s", attr1, attr2, attr1, status1KeyValue, attr2, status2KeyValue)
		}

		return nil
	}
}
