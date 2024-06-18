package task_management_worktype

import (
	"context"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/util/lists"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v131/platformclientv2"
)

/*
The resource_genesyscloud_task_management_worktype_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.
*/

type worktypeConfig struct {
	resID             string
	name              string
	description       string
	divisionId        string
	statuses          []worktypeStatusConfig
	defaultStatusName string
	defaultWorkbinId  string

	defaultDurationS    int
	defaultExpirationS  int
	defaultDueDurationS int
	defaultPriority     int
	defaultTtlS         int

	defaultLanguageId string
	defaultQueueId    string
	defaultSkillIds   []string
	assignmentEnabled bool

	schemaId      string
	schemaVersion int
}

type worktypeStatusConfig struct {
	id                           string
	name                         string
	description                  string
	category                     string
	destinationStatusNames       []string
	defaultDestinationStatusName string
	transitionDelay              int
	statusTransitionTime         string
}

// getWorktypeCreateFromResourceData maps data from schema ResourceData object to a platformclientv2.Worktypecreate
func getWorktypecreateFromResourceData(d *schema.ResourceData) platformclientv2.Worktypecreate {
	worktype := platformclientv2.Worktypecreate{
		Name:                         platformclientv2.String(d.Get("name").(string)),
		DivisionId:                   platformclientv2.String(d.Get("division_id").(string)),
		Description:                  platformclientv2.String(d.Get("description").(string)),
		DisableDefaultStatusCreation: platformclientv2.Bool(true),
		DefaultWorkbinId:             platformclientv2.String(d.Get("default_workbin_id").(string)),
		SchemaId:                     platformclientv2.String(d.Get("schema_id").(string)),
		SchemaVersion:                resourcedata.GetNillableValue[int](d, "schema_version"),

		DefaultPriority: platformclientv2.Int(d.Get("default_priority").(int)),

		DefaultLanguageId: resourcedata.GetNillableValue[string](d, "default_language_id"),
		DefaultQueueId:    resourcedata.GetNillableValue[string](d, "default_queue_id"),
		DefaultSkillIds:   lists.BuildSdkStringListFromInterfaceArray(d, "default_skills_ids"),
		AssignmentEnabled: platformclientv2.Bool(d.Get("assignment_enabled").(bool)),

		DefaultDurationSeconds:    resourcedata.GetNillableValue[int](d, "default_duration_seconds"),
		DefaultExpirationSeconds:  resourcedata.GetNillableValue[int](d, "default_expiration_seconds"),
		DefaultDueDurationSeconds: resourcedata.GetNillableValue[int](d, "default_due_duration_seconds"),
		DefaultTtlSeconds:         resourcedata.GetNillableValue[int](d, "default_ttl_seconds"),
	}

	return worktype
}

// getWorktypeupdateFromResourceData maps data from schema ResourceData object to a platformclientv2.Worktypeupdate
func getWorktypeupdateFromResourceData(d *schema.ResourceData, statuses *[]platformclientv2.Workitemstatus) platformclientv2.Worktypeupdate {

	worktype := platformclientv2.Worktypeupdate{}
	worktype.SetField("Name", platformclientv2.String(d.Get("name").(string)))
	if d.HasChange("description") {
		worktype.SetField("Description", platformclientv2.String(d.Get("description").(string)))
	}
	if d.HasChange("default_workbin_id") {
		worktype.SetField("DefaultWorkbinId", platformclientv2.String(d.Get("default_workbin_id").(string)))
	}

	if d.HasChange("default_priority") {
		worktype.SetField("DefaultPriority", platformclientv2.Int(d.Get("default_priority").(int)))
	}

	if d.HasChange("schema_id") {
		worktype.SetField("SchemaId", platformclientv2.String(d.Get("schema_id").(string)))
	}

	if d.HasChange("default_language_id") {
		worktype.SetField("DefaultLanguageId", resourcedata.GetNillableValue[string](d, "default_language_id"))
	}

	if d.HasChange("default_queue_id") {
		worktype.SetField("DefaultQueueId", resourcedata.GetNillableValue[string](d, "default_queue_id"))
	}

	if d.HasChange("default_skills_ids") {
		worktype.SetField("DefaultSkillIds", lists.BuildSdkStringListFromInterfaceArray(d, "default_skills_ids"))
	}

	if d.HasChange("assignment_enabled") {
		worktype.SetField("AssignmentEnabled", platformclientv2.Bool(d.Get("assignment_enabled").(bool)))
	}

	if d.HasChange("schema_version") {
		worktype.SetField("SchemaVersion", resourcedata.GetNillableValue[int](d, "schema_version"))
	}

	if d.HasChange("default_duration_seconds") {
		worktype.SetField("DefaultDurationSeconds", resourcedata.GetNillableValue[int](d, "default_duration_seconds"))
	}
	if d.HasChange("default_expiration_seconds") {
		worktype.SetField("DefaultExpirationSeconds", resourcedata.GetNillableValue[int](d, "default_duration_seconds"))
	}
	if d.HasChange("default_due_duration_seconds") {
		worktype.SetField("DefaultDueDurationSeconds", resourcedata.GetNillableValue[int](d, "default_due_duration_seconds"))
	}
	if d.HasChange("default_ttl_seconds") {
		worktype.SetField("DefaultTtlSeconds", resourcedata.GetNillableValue[int](d, "default_ttl_seconds"))
	}

	return worktype
}

// getWorktypeupdateFromResourceData maps data from schema ResourceData object to a platformclientv2.Worktypeupdate
func getWorktypeupdateFromResourceDataStatus(d *schema.ResourceData, statuses *[]platformclientv2.Workitemstatus) platformclientv2.Worktypeupdate {

	worktype := platformclientv2.Worktypeupdate{}
	worktype.SetField("Name", platformclientv2.String(d.Get("name").(string)))
	if d.HasChange("default_status_name") {
		worktype.SetField("DefaultStatusId", getStatusIdFromName(d.Get("default_status_name").(string), statuses))
	}
	return worktype
}

// getStatusFromName gets a platformclientv2.Workitemstatus from a  *[]platformclientv2.Workitemstatu by name
func getStatusFromName(statusName string, statuses *[]platformclientv2.Workitemstatus) *platformclientv2.Workitemstatus {
	if statuses == nil {
		return nil
	}

	for _, apiStatus := range *statuses {
		if statusName == *apiStatus.Name {
			return &apiStatus
		}
	}

	return nil
}

// getStatusIdFromName gets a status id from a  *[]platformclientv2.Workitemstatu by status name
func getStatusIdFromName(statusName string, statuses *[]platformclientv2.Workitemstatus) *string {
	if statuses == nil {
		return nil
	}

	for _, apiStatus := range *statuses {
		if statusName == *apiStatus.Name {
			return apiStatus.Id
		}
	}

	return nil
}

// getStatusIdFromName gets the status name from a  *[]platformclientv2.Workitemstatus by status id
func getStatusNameFromId(statusId string, statuses *[]platformclientv2.Workitemstatus) *string {
	if statuses == nil {
		return nil
	}

	for _, apiStatus := range *statuses {
		if statusId == *apiStatus.Id {
			return apiStatus.Name
		}
	}

	return nil
}

// buildWorkitemStatusCreates maps an []interface{} into a Genesys Cloud *[]platformclientv2.Workitemstatuscreate
func buildWorkitemStatusCreates(workitemStatuses []interface{}) *[]platformclientv2.Workitemstatuscreate {
	workitemStatusesSlice := make([]platformclientv2.Workitemstatuscreate, 0)

	for _, workitemStatus := range workitemStatuses {
		var sdkWorkitemStatus platformclientv2.Workitemstatuscreate
		workitemStatusMap, ok := workitemStatus.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkWorkitemStatus.Name, workitemStatusMap, "name")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkWorkitemStatus.Category, workitemStatusMap, "category")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkWorkitemStatus.Description, workitemStatusMap, "description")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkWorkitemStatus.StatusTransitionTime, workitemStatusMap, "status_transition_time")
		if statusTransitionDelaySec, ok := workitemStatusMap["status_transition_delay_seconds"]; ok && statusTransitionDelaySec.(int) > 0 {
			sdkWorkitemStatus.StatusTransitionDelaySeconds = platformclientv2.Int(statusTransitionDelaySec.(int))
		}

		workitemStatusesSlice = append(workitemStatusesSlice, sdkWorkitemStatus)
	}

	return &workitemStatusesSlice
}

// buildWorkitemStatusUpdates maps an []interface{} into a Genesys Cloud *[]platformclientv2.Workitemstatusupdate
// workitemStatuses is the terraform resource object attribute while apiStatuses is the existing Genesys Cloud
// statuses to be used as reference for the 'destination status ids'
func buildWorkitemStatusUpdates(workitemStatuses []interface{}, apiStatuses *[]platformclientv2.Workitemstatus) *[]platformclientv2.Workitemstatusupdate {
	workitemStatussSlice := make([]platformclientv2.Workitemstatusupdate, 0)

	// Inner func get the status id from a status name.
	getStatusIdFromNameFn := func(statusName string) *string {
		return getStatusIdFromName(statusName, apiStatuses)
	}

	// Inner func for use in building the destination status ids from the status names defined
	buildStatusIdFn := func(statusNames []interface{}) *[]string {
		statusIds := []string{}
		for _, name := range statusNames {
			if statusId := getStatusIdFromName(name.(string), apiStatuses); statusId != nil {
				statusIds = append(statusIds, *statusId)
			}
		}

		return &statusIds
	}

	for _, workitemStatus := range workitemStatuses {
		sdkWorkitemStatus := platformclientv2.Workitemstatusupdate{}
		workitemStatusMap, ok := workitemStatus.(map[string]interface{})
		if !ok {
			continue
		}

		// For the following if some attributes are not provided in terraform file we
		// explicitly set the SDK attribute to nil to nullify its value in the API

		if name, ok := workitemStatusMap["name"]; ok {
			sdkWorkitemStatus.SetField("Name", platformclientv2.String(name.(string)))
		}

		if description, ok := workitemStatusMap["description"]; ok {
			sdkWorkitemStatus.SetField("Description", platformclientv2.String(description.(string)))
		}

		if statusTransitionTime, ok := workitemStatusMap["status_transition_time"]; ok {
			sdkWorkitemStatus.SetField("StatusTransitionTime", platformclientv2.String(statusTransitionTime.(string)))
		} else {
			sdkWorkitemStatus.SetField("StatusTransitionTime", nil)
		}

		if statusTransitionDelaySec, ok := workitemStatusMap["status_transition_delay_seconds"]; ok && statusTransitionDelaySec.(int) > 0 {
			sdkWorkitemStatus.SetField("StatusTransitionDelaySeconds", platformclientv2.Int(statusTransitionDelaySec.(int)))
		} else {
			sdkWorkitemStatus.SetField("StatusTransitionDelaySeconds", nil)
		}

		if destinationStatuses, ok := workitemStatusMap["destination_status_names"]; ok {
			statusIds := buildStatusIdFn(destinationStatuses.([]interface{}))
			sdkWorkitemStatus.SetField("DestinationStatusIds", statusIds)
		} else {
			sdkWorkitemStatus.SetField("DestinationStatusIds", nil)
		}

		if defaultDestination, ok := workitemStatusMap["default_destination_status_name"]; ok {
			defaultDestStatusId := getStatusIdFromNameFn(defaultDestination.(string))
			sdkWorkitemStatus.SetField("DefaultDestinationStatusId", defaultDestStatusId)
		} else {
			sdkWorkitemStatus.SetField("DefaultDestinationStatusId", nil)
		}

		workitemStatussSlice = append(workitemStatussSlice, sdkWorkitemStatus)
	}

	return &workitemStatussSlice
}

// flattenWorkitemStatusReferences maps a Genesys Cloud *[]platformclientv2.Workitemstatusreference into a []interface{}
// Sadly the API only returns the ID in the ref object (even if the name is defined in the API model), so we still need the
// existingStatuses parameter to get the name for resolving back into resource data
func flattenWorkitemStatusReferences(workitemStatusReferences *[]platformclientv2.Workitemstatusreference, existingStatuses *[]platformclientv2.Workitemstatus) []interface{} {
	if len(*workitemStatusReferences) == 0 {
		return nil
	}

	var workitemStatusReferenceList []interface{}
	for _, workitemStatusReference := range *workitemStatusReferences {
		for _, existingStatus := range *existingStatuses {
			if *workitemStatusReference.Id == *existingStatus.Id {
				workitemStatusReferenceList = append(workitemStatusReferenceList, existingStatus.Name)
			}
		}
	}

	return workitemStatusReferenceList
}

// flattenWorkitemStatuss maps a Genesys Cloud *[]platformclientv2.Workitemstatus into a []interface{}
func flattenWorkitemStatuses(workitemStatuses *[]platformclientv2.Workitemstatus) []interface{} {
	if len(*workitemStatuses) == 0 {
		return nil
	}

	// Containing function for flattening because we need to use
	// worktype statuses as reference for the method
	flattenStatusRefsWithExisting := func(refs *[]platformclientv2.Workitemstatusreference) []interface{} {
		return flattenWorkitemStatusReferences(refs, workitemStatuses)
	}

	var workitemStatusList []interface{}
	for _, workitemStatus := range *workitemStatuses {
		workitemStatusMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(workitemStatusMap, "id", workitemStatus.Id)
		resourcedata.SetMapValueIfNotNil(workitemStatusMap, "name", workitemStatus.Name)
		resourcedata.SetMapValueIfNotNil(workitemStatusMap, "category", workitemStatus.Category)
		resourcedata.SetMapValueIfNotNil(workitemStatusMap, "description", workitemStatus.Description)
		resourcedata.SetMapValueIfNotNil(workitemStatusMap, "status_transition_delay_seconds", workitemStatus.StatusTransitionDelaySeconds)
		resourcedata.SetMapValueIfNotNil(workitemStatusMap, "status_transition_time", workitemStatus.StatusTransitionTime)

		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(workitemStatusMap, "destination_status_names", workitemStatus.DestinationStatuses, flattenStatusRefsWithExisting)
		if workitemStatus.DefaultDestinationStatus != nil {
			resourcedata.SetMapValueIfNotNil(workitemStatusMap, "default_destination_status_name", getStatusNameFromId(*workitemStatus.DefaultDestinationStatus.Id, workitemStatuses))
		}

		workitemStatusList = append(workitemStatusList, workitemStatusMap)
	}

	return workitemStatusList
}

// flattenRoutingSkillReferences maps a Genesys Cloud *[]platformclientv2.Routingskillreference into a []interface{}
func flattenRoutingSkillReferences(routingSkillReferences *[]platformclientv2.Routingskillreference) []interface{} {
	if len(*routingSkillReferences) == 0 {
		return nil
	}

	var routingSkillReferenceList []interface{}
	for _, routingSkillReference := range *routingSkillReferences {
		routingSkillReferenceList = append(routingSkillReferenceList, *routingSkillReference.Id)
	}

	return routingSkillReferenceList
}

// getStatusesForUpdateAndCreation takes a resource data list []interface{} of statuses and determines if they
// are to be created or just updated. Returns the two lists.
func getStatusesForUpdateAndCreation(statuses []interface{}, existingStatuses *[]platformclientv2.Workitemstatus) (forCreation []interface{}, forUpdate []interface{}) {
	forCreation = make([]interface{}, 0)
	forUpdate = make([]interface{}, 0)

	// We will consider it the same status and update in-place if the name and the category matches.
	// else, a new status will be created.
	for _, status := range statuses {
		statusMap := status.(map[string]interface{})
		statusName, ok := statusMap["name"]
		if !ok {
			continue
		}
		statusCat, ok := statusMap["category"]
		if !ok {
			continue
		}
		toCreateNewStatus := true

		// If the status matches an existing name and same category then we'll consider
		// it the same and not create a new one.
		for _, existingStatus := range *existingStatuses {
			if *existingStatus.Name == statusName && *existingStatus.Category == statusCat {
				toCreateNewStatus = false
				break
			}
		}

		if toCreateNewStatus {
			forCreation = append(forCreation, status)
		} else {
			forUpdate = append(forUpdate, status)
		}
	}

	return forCreation, forUpdate
}

// createWorktypeStatuses creates new statuses as defined in the config. This is just the initial
// creation as some statuses also need to be updated separately to build the destination status references.
func createWorktypeStatuses(ctx context.Context, proxy *taskManagementWorktypeProxy, worktypeId string, statuses []interface{}) (*[]platformclientv2.Workitemstatus, error) {
	ret := []platformclientv2.Workitemstatus{}

	sdkWorkitemStatusCreates := buildWorkitemStatusCreates(statuses)
	for _, statusCreate := range *sdkWorkitemStatusCreates {
		status, resp, err := proxy.createTaskManagementWorktypeStatus(ctx, worktypeId, &statusCreate)
		if err != nil {
			return nil, fmt.Errorf("failed to create worktype status %s: %v %v", *statusCreate.Name, err, resp)
		}

		ret = append(ret, *status)
	}

	return &ret, nil
}

// updateWorktypeStatuses updates the statuses of a worktype. There are two modes depending if the passed statuses
// is newly created or not. For newly created, we just check if there's any need to resolve references since they still
// have none. For existing statuses, they should be passed as already validated (has change - because API will return error
// if there's no change to the status), the method will not check it.
func updateWorktypeStatuses(ctx context.Context, proxy *taskManagementWorktypeProxy, worktypeId string, statuses []interface{}, isNewlyCreated bool) (*[]platformclientv2.Workitemstatus, error) {
	ret := []platformclientv2.Workitemstatus{}

	// Get all the worktype statuses so we'll have the new statuses for referencing
	worktype, resp, err := proxy.getTaskManagementWorktypeById(ctx, worktypeId)
	if err != nil {
		return nil, fmt.Errorf("failed to get task management worktype %s: %v %v", worktypeId, err, resp)
	}

	// Update the worktype statuses as they need to build the "destination status" references
	sdkWorkitemStatusUpdates := buildWorkitemStatusUpdates(statuses, worktype.Statuses)
	for _, statusUpdate := range *sdkWorkitemStatusUpdates {
		existingStatus := getStatusFromName(*statusUpdate.Name, worktype.Statuses)
		if existingStatus.Id == nil {
			return nil, fmt.Errorf("failed to update a status %s. Not found in the worktype %s: %v", *statusUpdate.Name, *worktype.Name, err)
		}

		// API does not allow updating a status with no actual change.
		// For newly created statuses, update portion is only for resolving status references, so skip statuses where
		// "destination statuses" and "default destination id" are not set.
		if isNewlyCreated &&
			(statusUpdate.DefaultDestinationStatusId == nil || *statusUpdate.DefaultDestinationStatusId == "") &&
			(statusUpdate.DestinationStatusIds == nil || len(*statusUpdate.DestinationStatusIds) == 0) {
			continue
		}

		status, resp, err := proxy.updateTaskManagementWorktypeStatus(ctx, *worktype.Id, *existingStatus.Id, &statusUpdate)
		if err != nil {
			return nil, fmt.Errorf("failed to update worktype status %s: %v %v", *statusUpdate.Name, err, resp)
		}

		ret = append(ret, *status)
	}

	return &ret, nil
}

// updateDefaultStatusName updates a worktype's default status name. This should be called after
// the statuses and their references have been finalized.
func updateDefaultStatusName(ctx context.Context, proxy *taskManagementWorktypeProxy, d *schema.ResourceData, worktypeId string) error {
	worktype, resp, err := proxy.getTaskManagementWorktypeById(ctx, worktypeId)
	if err != nil {
		return fmt.Errorf("failed to get task management worktype: %s", err)
	}

	taskManagementWorktype := getWorktypeupdateFromResourceDataStatus(d, worktype.Statuses)
	_, resp, err = proxy.updateTaskManagementWorktype(ctx, *worktype.Id, &taskManagementWorktype)
	if err != nil {
		return fmt.Errorf("failed to update worktype's default status name %s %v", err, resp)
	}

	return nil
}

// getStatusIdFromName gets the status id from name for the test util struct worktypeConfig
func (wt *worktypeConfig) getStatusIdFromName(name string) *string {
	for _, s := range wt.statuses {
		if s.name == name {
			return &s.id
		}
	}

	return nil
}

// GenerateWorktypeResourceBasic generates a terraform config string for a basic worktype
func GenerateWorktypeResourceBasic(resId, name, description, workbinResourceId, schemaResourceId, attrs string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		name = "%s"
		description = "%s"
		default_workbin_id = %s
		schema_id = %s
		%s
	}
	`, resourceName, resId, name, description, workbinResourceId, schemaResourceId, attrs)
}
