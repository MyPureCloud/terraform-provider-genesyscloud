package task_management_worktype_status

import (
	"context"
	"fmt"
	"strings"

	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"
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
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update worktype %s with default status %s.", worktypeId, statusId), resp)
	}

	return nil
}

func GenerateWorktypeStatusResource(
	resourceLabel,
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
`, resourceLabel, workTypeId, name, category, description, defaultDestinationStatusId, statusTransitionTime, strings.Join(attrs, "\n"))
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

// WorktypeStatusRefResolver resolves a bare status ID to a proper Terraform reference
// by searching the worktype_status SanitizedResourceMap for a composite key ending with "/<statusId>".
// This is needed because the worktype_status resource uses composite IDs (worktypeId/statusId)
// as map keys, but other resources store only the bare statusId in their state.
func WorktypeStatusRefResolver(attrName string) func(configMap map[string]interface{}, exporters map[string]*resourceExporter.ResourceExporter, resourceLabel string) error {
	return func(configMap map[string]interface{}, exporters map[string]*resourceExporter.ResourceExporter, resourceLabel string) error {
		statusIdRaw, ok := configMap[attrName]
		if !ok || statusIdRaw == nil {
			return nil
		}

		statusId, ok := statusIdRaw.(string)
		if !ok || statusId == "" {
			return nil
		}

		// If already resolved to a reference, skip
		if strings.HasPrefix(statusId, "${") {
			return nil
		}

		exporter := exporters[ResourceType]
		if exporter == nil {
			return nil
		}

		idMetaMap := exporter.SanitizedResourceMap
		if idMetaMap == nil {
			return nil
		}

		// Search for a composite key that ends with "/<statusId>"
		suffix := "/" + statusId
		for compositeId, meta := range idMetaMap {
			if strings.HasSuffix(compositeId, suffix) && meta != nil && meta.BlockLabel != "" {
				configMap[attrName] = fmt.Sprintf("${%s.%s.id}", ResourceType, meta.BlockLabel)
				return nil
			}
		}

		return nil
	}
}

// WorktypeStatusArrayRefResolver resolves bare status IDs in a list attribute to proper Terraform references.
// Used for attributes like destination_status_ids which are arrays of status IDs.
func WorktypeStatusArrayRefResolver(attrName string) func(configMap map[string]interface{}, exporters map[string]*resourceExporter.ResourceExporter, resourceLabel string) error {
	return func(configMap map[string]interface{}, exporters map[string]*resourceExporter.ResourceExporter, resourceLabel string) error {
		arrRaw, ok := configMap[attrName]
		if !ok || arrRaw == nil {
			return nil
		}

		arr, ok := arrRaw.([]interface{})
		if !ok || len(arr) == 0 {
			return nil
		}

		exporter := exporters[ResourceType]
		if exporter == nil {
			return nil
		}

		idMetaMap := exporter.SanitizedResourceMap
		if idMetaMap == nil {
			return nil
		}

		resolved := make([]interface{}, 0, len(arr))
		for _, item := range arr {
			statusId, ok := item.(string)
			if !ok || statusId == "" {
				resolved = append(resolved, item)
				continue
			}

			// If already resolved to a reference, keep it
			if strings.HasPrefix(statusId, "${") {
				resolved = append(resolved, statusId)
				continue
			}

			// Search for a composite key that ends with "/<statusId>"
			found := false
			suffix := "/" + statusId
			for compositeId, meta := range idMetaMap {
				if strings.HasSuffix(compositeId, suffix) && meta != nil && meta.BlockLabel != "" {
					resolved = append(resolved, fmt.Sprintf("${%s.%s.id}", ResourceType, meta.BlockLabel))
					found = true
					break
				}
			}

			if !found {
				resolved = append(resolved, statusId)
			}
		}

		configMap[attrName] = resolved
		return nil
	}
}
