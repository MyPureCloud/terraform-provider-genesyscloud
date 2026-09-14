package task_management_worktype_status_transition

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
)

const noChangeForRecordError = "No change for the record is obtained"

// ModifyStatusIdStateValue will change the statusId before it is saved in the state file.
// The worktype_status resource saves the status id as <worktypeId>/<statusId>.
// We only want to save the statusId in the state as the api will only return the status id
// and this cause would 'plan not empty' if we save the id as <worktypeId>/<statusId>
func modifyStatusIdStateValue(id interface{}) string {
	statusId := id.(string)
	idWithoutSuffix := strings.TrimSuffix(statusId, " transition")
	if strings.Contains(idWithoutSuffix, "/") {
		return strings.Split(idWithoutSuffix, "/")[1]
	}

	return statusId
}

// splitWorktypeStatusTerraformTransitionId will split the status resource id which is in the form
// <worktypeId>/<statusId> into just the worktypeId and statusId string
func splitWorktypeStatusTerraformTransitionId(id string) (worktypeId string, statusId string) {
	idWithoutSuffix := strings.TrimSuffix(id, " transition")
	if strings.Contains(idWithoutSuffix, "/") {
		return strings.Split(idWithoutSuffix, "/")[0], strings.Split(idWithoutSuffix, "/")[1]
	} else {
		return "", idWithoutSuffix
	}
}

func fetchWorktypeStatusTerraformId(id string) (statusId string) {
	idWithoutSuffix := strings.TrimSuffix(id, " transition")
	if strings.Contains(idWithoutSuffix, "/") {
		return strings.Split(idWithoutSuffix, "/")[1]
	} else {
		return idWithoutSuffix
	}
}

func fetchWorktypeStatusId(id string) (statusId string) {
	idWithoutSuffix := strings.TrimSuffix(id, " transition")
	fmt.Println(idWithoutSuffix)
	return strings.Split(idWithoutSuffix, "/")[1]
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

func updateWorktypeDefaultStatus(ctx context.Context, proxy *taskManagementWorktypeStatusTransitionProxy, worktypeId string, statusId string) diag.Diagnostics {
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

func GenerateWorkTypeStatusResourceTransition(
	resourceLabel,
	workTypeId,
	statusId,
	defaultDestinationStatusId string,
	destinationStatusId string,
	delaySeconds string,
	attrs ...string,
) string {
	return fmt.Sprintf(
		`resource "genesyscloud_task_management_worktype_status_transition" "%s" {
		worktype_id = %s
		status_id = %s
		default_destination_status_id = %s
		destination_status_ids = [%s]
status_transition_delay_seconds = "%s"
		%s
	}
`, resourceLabel, workTypeId, statusId, defaultDestinationStatusId, destinationStatusId, delaySeconds, strings.Join(attrs, "\n"))
}

func GenerateWorkTypeStatusResourceTransitionWithoutAutoTransition(
	resourceLabel,
	workTypeId,
	statusId,
	destinationStatusId string,
	attrs ...string,
) string {
	return fmt.Sprintf(
		`resource "genesyscloud_task_management_worktype_status_transition" "%s" {
		worktype_id = %s
		status_id = %s
		destination_status_ids = [%s]
		%s
	}
`, resourceLabel, workTypeId, statusId, destinationStatusId, strings.Join(attrs, "\n"))
}

func buildWorkitemStatusTransitionPatch(d *schema.ResourceData, workitemStatus *platformclientv2.Workitemstatus, destinationStatusIds *[]string, defaultDestinationStatusId *string, isCreate bool) *Workitemstatusupdate {
	body := &Workitemstatusupdate{}
	body.SetField("Name", workitemStatus.Name)
	body.SetField("Description", workitemStatus.Description)
	body.SetField("DestinationStatusIds", destinationStatusIds)

	if defaultDestinationStatusId != nil && *defaultDestinationStatusId != "" {
		body.SetField("DefaultDestinationStatusId", defaultDestinationStatusId)
	} else {
		setOptionalStringPatchField(d, body, "default_destination_status_id", "DefaultDestinationStatusId", isCreate)
	}

	setOptionalIntPatchField(d, body, "status_transition_delay_seconds", "StatusTransitionDelaySeconds", isCreate)
	setOptionalStringPatchField(d, body, "status_transition_time", "StatusTransitionTime", isCreate)

	return body
}

func setOptionalStringPatchField(d *schema.ResourceData, body *Workitemstatusupdate, key string, fieldName string, isCreate bool) {
	if value, ok := d.GetOk(key); ok {
		v := value.(string)
		if v != "" {
			body.SetField(fieldName, &v)
			return
		}
	}

	// On update, always send JSON null so optional fields can be cleared. Gating on
	// HasChange misses unsets when StateFunc normalizes status IDs.
	if !isCreate {
		body.SetField(fieldName, nil)
	}
}

func setOptionalIntPatchField(d *schema.ResourceData, body *Workitemstatusupdate, key string, fieldName string, isCreate bool) {
	if value, ok := d.GetOk(key); ok {
		v := value.(int)
		body.SetField(fieldName, &v)
		return
	}

	if !isCreate {
		body.SetField(fieldName, nil)
	}
}

func isNoChangeForRecordError(err error, resp *platformclientv2.APIResponse) bool {
	if resp != nil && strings.Contains(resp.ErrorMessage, noChangeForRecordError) {
		return true
	}
	return err != nil && strings.Contains(err.Error(), noChangeForRecordError)
}

// destinationStatusIdsOmittedFromConfig reports whether destination_status_ids is absent or
// empty in Terraform config. An empty list means the Workitem can transition to all other
// statuses; the API then returns the expanded list, which must not be written back to state.
func destinationStatusIdsOmittedFromConfig(d *schema.ResourceData) bool {
	if d == nil {
		return false
	}
	raw := d.GetRawConfig()
	if !raw.IsKnown() || raw.IsNull() || !raw.CanIterateElements() {
		return false
	}
	attr := raw.GetAttr("destination_status_ids")
	if !attr.IsKnown() {
		return false
	}
	if attr.IsNull() {
		return true
	}
	if attr.Type().IsListType() || attr.Type().IsSetType() || attr.Type().IsTupleType() {
		return attr.LengthInt() == 0
	}
	return false
}

func setDestinationStatusIdsState(d *schema.ResourceData, workitemStatus *platformclientv2.Workitemstatus, importing bool) {
	// Import reads should populate destinations from the API. Afterwards, an empty
	// destination_status_ids in config/state means "all other statuses"; do not write
	// the API-expanded list back or terraform plan never settles.
	if !importing {
		if destinationStatusIdsOmittedFromConfig(d) {
			_ = d.Set("destination_status_ids", []interface{}{})
			return
		}
		if dest, ok := d.GetOk("destination_status_ids"); !ok || len(dest.([]interface{})) == 0 {
			_ = d.Set("destination_status_ids", []interface{}{})
			return
		}
	}
	if workitemStatus == nil || workitemStatus.DestinationStatuses == nil {
		return
	}
	destinationStatuses := make([]interface{}, len(*workitemStatus.DestinationStatuses))
	for i, v := range *workitemStatus.DestinationStatuses {
		if v.Id != nil {
			destinationStatuses[i] = *v.Id
		}
	}
	_ = d.Set("destination_status_ids", destinationStatuses)
}

func normalizeStatusIdList(ids []string) []string {
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = fetchWorktypeStatusTerraformId(id)
	}
	return out
}

func destinationStatusIdsFromWorkitemStatus(status *platformclientv2.Workitemstatus) []string {
	if status == nil || status.DestinationStatuses == nil {
		return []string{}
	}
	ids := make([]string, 0, len(*status.DestinationStatuses))
	for _, s := range *status.DestinationStatuses {
		if s.Id != nil {
			ids = append(ids, *s.Id)
		}
	}
	return ids
}

// workitemStatusTransitionMatchesConfig reports whether API state already matches the
// Terraform config. An empty destination_status_ids list means "all other statuses", so
// it is not compared against the expanded API list.
func workitemStatusTransitionMatchesConfig(d *schema.ResourceData, status *platformclientv2.Workitemstatus) bool {
	if d == nil || status == nil {
		return false
	}

	desiredDefault := resourcedata.GetNillableValue[string](d, "default_destination_status_id")
	if desiredDefault != nil && *desiredDefault != "" {
		id := fetchWorktypeStatusTerraformId(*desiredDefault)
		if status.DefaultDestinationStatus == nil || status.DefaultDestinationStatus.Id == nil || *status.DefaultDestinationStatus.Id != id {
			return false
		}
	} else if status.DefaultDestinationStatus != nil && status.DefaultDestinationStatus.Id != nil && *status.DefaultDestinationStatus.Id != "" {
		return false
	}

	if delay, ok := d.GetOk("status_transition_delay_seconds"); ok {
		desired := delay.(int)
		if status.StatusTransitionDelaySeconds == nil || *status.StatusTransitionDelaySeconds != desired {
			return false
		}
	} else if status.StatusTransitionDelaySeconds != nil && *status.StatusTransitionDelaySeconds != 0 {
		return false
	}

	if timeVal, ok := d.GetOk("status_transition_time"); ok {
		desired := timeVal.(string)
		if desired != "" && (status.StatusTransitionTime == nil || *status.StatusTransitionTime != desired) {
			return false
		}
	} else if status.StatusTransitionTime != nil && *status.StatusTransitionTime != "" {
		return false
	}

	desiredDest := lists.BuildSdkStringListFromInterfaceArray(d, "destination_status_ids")
	if desiredDest != nil && len(*desiredDest) > 0 {
		if !lists.AreEquivalent(normalizeStatusIdList(*desiredDest), destinationStatusIdsFromWorkitemStatus(status)) {
			return false
		}
	}

	return true
}

func GenerateWorktypeStatusResourceWithDependsOn(
	resourceLabel,
	workTypeId,
	name,
	category,
	description string,
	defaultDestinationStatusId string,
	statusTransitionTime string,
	dependsOn string,
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
	depends_on = [%s]
		%s
	}
`, resourceLabel, workTypeId, name, category, description, defaultDestinationStatusId, statusTransitionTime, dependsOn, strings.Join(attrs, "\n"))
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
			_, status1KeyValue = splitWorktypeStatusTerraformTransitionId(status1KeyValue)
		}

		status2KeyValue := status2.Primary.Attributes[key2]
		if strings.Contains(status2KeyValue, "/") {
			_, status2KeyValue = splitWorktypeStatusTerraformTransitionId(status2KeyValue)
		}

		if status1KeyValue != status2KeyValue {
			attr1 := statusResource1 + "." + key1
			attr2 := statusResource2 + "." + key2
			return fmt.Errorf("%s not equal to %s\n %s = %s\n %s = %s", attr1, attr2, attr1, status1KeyValue, attr2, status2KeyValue)
		}

		return nil
	}
}
