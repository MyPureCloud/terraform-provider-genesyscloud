package task_management_worktype

import (
	"terraform-provider-genesyscloud/genesyscloud/util/lists"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

/*
The resource_genesyscloud_task_management_worktype_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.
*/

// getWorktypeCreateFromResourceData maps data from schema ResourceData object to a platformclientv2.Worktypecreate
func getWorktypecreateFromResourceData(d *schema.ResourceData) platformclientv2.Worktypecreate {
	worktype := platformclientv2.Worktypecreate{
		Name:                         platformclientv2.String(d.Get("name").(string)),
		DivisionId:                   platformclientv2.String(d.Get("division_id").(string)),
		Description:                  platformclientv2.String(d.Get("description").(string)),
		DisableDefaultStatusCreation: platformclientv2.Bool(true),
		DefaultWorkbinId:             platformclientv2.String(d.Get("default_workbin_id").(string)),

		DefaultPriority: platformclientv2.Int(d.Get("default_priority").(int)),

		DefaultLanguageId: platformclientv2.String(d.Get("default_language_id").(string)),
		DefaultQueueId:    platformclientv2.String(d.Get("default_queue_id").(string)),
		DefaultSkillIds:   lists.BuildSdkStringListFromInterfaceArray(d, "default_skills_ids"),
		AssignmentEnabled: platformclientv2.Bool(d.Get("assignment_enabled").(bool)),
		SchemaId:          platformclientv2.String(d.Get("schema_id").(string)),
	}

	// For the following we want the 0 value, but also nil (which has a different default value set by API)
	if d.Get("default_duration_seconds") != nil {
		worktype.DefaultDurationSeconds = platformclientv2.Int(d.Get("default_duration_seconds").(int))
	}

	if d.Get("default_expiration_seconds") != nil {
		worktype.DefaultExpirationSeconds = platformclientv2.Int(d.Get("default_expiration_seconds").(int))
	}

	if d.Get("default_due_duration_seconds") != nil {
		worktype.DefaultDueDurationSeconds = platformclientv2.Int(d.Get("default_due_duration_seconds").(int))
	}

	if d.Get("default_ttl_seconds") != nil {
		worktype.DefaultTtlSeconds = platformclientv2.Int(d.Get("default_ttl_seconds").(int))
	}

	return worktype
}

// getWorktypeupdateFromResourceData maps data from schema ResourceData object to a platformclientv2.Worktypeupdate
func getWorktypeupdateFromResourceData(d *schema.ResourceData) platformclientv2.Worktypeupdate {
	worktype := platformclientv2.Worktypeupdate{
		Name:             platformclientv2.String(d.Get("name").(string)),
		Description:      platformclientv2.String(d.Get("description").(string)),
		DefaultWorkbinId: platformclientv2.String(d.Get("default_workbin_id").(string)),

		DefaultPriority: platformclientv2.Int(d.Get("default_priority").(int)),

		DefaultLanguageId: platformclientv2.String(d.Get("default_language_id").(string)),
		DefaultQueueId:    platformclientv2.String(d.Get("default_queue_id").(string)),
		DefaultSkillIds:   lists.BuildSdkStringListFromInterfaceArray(d, "default_skills_ids"),
		AssignmentEnabled: platformclientv2.Bool(d.Get("assignment_enabled").(bool)),
		SchemaId:          platformclientv2.String(d.Get("schema_id").(string)),
	}

	// For the following we want the 0 value, but also nil (which has a different default value set by API)
	if d.Get("default_duration_seconds") != nil {
		worktype.DefaultDurationSeconds = platformclientv2.Int(d.Get("default_duration_seconds").(int))
	}

	if d.Get("default_expiration_seconds") != nil {
		worktype.DefaultExpirationSeconds = platformclientv2.Int(d.Get("default_expiration_seconds").(int))
	}

	if d.Get("default_due_duration_seconds") != nil {
		worktype.DefaultDueDurationSeconds = platformclientv2.Int(d.Get("default_due_duration_seconds").(int))
	}

	if d.Get("default_ttl_seconds") != nil {
		worktype.DefaultTtlSeconds = platformclientv2.Int(d.Get("default_ttl_seconds").(int))
	}

	return worktype
}

// buildLocalTime converts a local time []interface{} into a Genesys Cloud *platformclientv2.Localtime
func buildLocalTime(localTimes []interface{}) *platformclientv2.Localtime {
	localTimesSlice := make([]platformclientv2.Localtime, 0)
	for _, localTime := range localTimes {
		var sdkLocalTime platformclientv2.Localtime
		localTimesMap, ok := localTime.(map[string]interface{})
		if !ok {
			continue
		}

		sdkLocalTime.Hour = platformclientv2.Int(localTimesMap["hour"].(int))
		sdkLocalTime.Minute = platformclientv2.Int(localTimesMap["minute"].(int))
		sdkLocalTime.Second = platformclientv2.Int(localTimesMap["second"].(int))

		localTimesSlice = append(localTimesSlice, sdkLocalTime)
	}

	return &(localTimesSlice[0])
}

// buildWorkitemStatusCreates maps an []interface{} into a Genesys Cloud *[]platformclientv2.Workitemstatuscreate
func buildWorkitemStatusCreates(workitemStatuses []interface{}) *[]platformclientv2.Workitemstatuscreate {
	workitemStatussSlice := make([]platformclientv2.Workitemstatuscreate, 0)
	for _, workitemStatus := range workitemStatuses {
		var sdkWorkitemStatus platformclientv2.Workitemstatuscreate
		workitemStatussMap, ok := workitemStatus.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkWorkitemStatus.Name, workitemStatussMap, "name")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkWorkitemStatus.Category, workitemStatussMap, "category")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkWorkitemStatus.Description, workitemStatussMap, "description")
		sdkWorkitemStatus.StatusTransitionDelaySeconds = platformclientv2.Int(workitemStatussMap["status_transition_delay_seconds"].(int))
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkWorkitemStatus.StatusTransitionTime, workitemStatussMap, "status_transition_time", buildLocalTime)

		workitemStatussSlice = append(workitemStatussSlice, sdkWorkitemStatus)
	}

	return &workitemStatussSlice
}

// getStatusIdFromName gets a status id from a  *[]platformclientv2.Workitemstatu by status name
func getStatusIdFromName(statusName string, statuses *[]platformclientv2.Workitemstatus) *string {
	for _, apiStatus := range *statuses {
		if statusName == *apiStatus.Name {
			return &statusName
		}
	}

	return nil
}

// getStatusIdFromName gets the status name from a  *[]platformclientv2.Workitemstatu by status id
func getStatusNameFromId(statusId string, statuses *[]platformclientv2.Workitemstatus) *string {
	for _, apiStatus := range *statuses {
		if statusId == *apiStatus.Id {
			return &statusId
		}
	}

	return nil
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
		var sdkWorkitemStatus platformclientv2.Workitemstatusupdate
		workitemStatussMap, ok := workitemStatus.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkWorkitemStatus.Name, workitemStatussMap, "name")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkWorkitemStatus.Description, workitemStatussMap, "description")
		sdkWorkitemStatus.StatusTransitionDelaySeconds = platformclientv2.Int(workitemStatussMap["status_transition_delay_seconds"].(int))
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkWorkitemStatus.StatusTransitionTime, workitemStatussMap, "status_transition_time", buildLocalTime)

		// Destination Statuses
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkWorkitemStatus.DestinationStatusIds, workitemStatussMap, "destination_status_names", buildStatusIdFn)
		resourcedata.BuildSDKStringValueIfNotNilTransform(&sdkWorkitemStatus.DefaultDestinationStatusId, workitemStatussMap, "default_destination_status_name", getStatusIdFromNameFn)

		workitemStatussSlice = append(workitemStatussSlice, sdkWorkitemStatus)
	}

	return &workitemStatussSlice
}

// buildWorkitemSchemas maps an []interface{} into a Genesys Cloud *[]platformclientv2.Workitemschema
func buildWorkitemSchemas(workitemSchemas []interface{}) *[]platformclientv2.Workitemschema {
	workitemSchemasSlice := make([]platformclientv2.Workitemschema, 0)
	for _, workitemSchema := range workitemSchemas {
		var sdkWorkitemSchema platformclientv2.Workitemschema
		workitemSchemasMap, ok := workitemSchema.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkWorkitemSchema.Name, workitemSchemasMap, "name")

		workitemSchemasSlice = append(workitemSchemasSlice, sdkWorkitemSchema)
	}

	return &workitemSchemasSlice
}

// flattenWorkitemStatusReferences maps a Genesys Cloud *[]platformclientv2.Workitemstatusreference into a []interface{}
func flattenWorkitemStatusReferences(workitemStatusReferences *[]platformclientv2.Workitemstatusreference) []interface{} {
	if len(*workitemStatusReferences) == 0 {
		return nil
	}

	var workitemStatusReferenceList []interface{}
	for _, workitemStatusReference := range *workitemStatusReferences {
		workitemStatusReferenceMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(workitemStatusReferenceMap, "name", workitemStatusReference.Name)

		workitemStatusReferenceList = append(workitemStatusReferenceList, workitemStatusReferenceMap)
	}

	return workitemStatusReferenceList
}

// flattenLocalTime maps a Genesys Cloud *[]platformclientv2.Localtime into a []interface{}
func flattenLocalTime(localTimes *[]platformclientv2.Localtime) []interface{} {
	if len(*localTimes) == 0 {
		return nil
	}

	var localTimeList []interface{}
	for _, localTime := range *localTimes {
		localTimeMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(localTimeMap, "hour", localTime.Hour)
		resourcedata.SetMapValueIfNotNil(localTimeMap, "minute", localTime.Minute)
		resourcedata.SetMapValueIfNotNil(localTimeMap, "second", localTime.Second)

		localTimeList = append(localTimeList, localTimeMap)
	}

	return localTimeList
}

// flattenWorkitemStatuss maps a Genesys Cloud *[]platformclientv2.Workitemstatus into a []interface{}
func flattenWorkitemStatuses(workitemStatuses *[]platformclientv2.Workitemstatus) []interface{} {
	if len(*workitemStatuses) == 0 {
		return nil
	}

	var workitemStatusList []interface{}
	for _, workitemStatus := range *workitemStatuses {
		workitemStatusMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(workitemStatusMap, "name", workitemStatus.Name)
		resourcedata.SetMapValueIfNotNil(workitemStatusMap, "category", workitemStatus.Category)
		resourcedata.SetMapValueIfNotNil(workitemStatusMap, "description", workitemStatus.Description)
		resourcedata.SetMapValueIfNotNil(workitemStatusMap, "status_transition_delay_seconds", workitemStatus.StatusTransitionDelaySeconds)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(workitemStatusMap, "status_transition_time", &[]platformclientv2.Localtime{*workitemStatus.StatusTransitionTime}, flattenLocalTime)

		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(workitemStatusMap, "destination_statuses", workitemStatus.DestinationStatuses, flattenWorkitemStatusReferences)
		if workitemStatus.DefaultDestinationStatus != nil {
			resourcedata.SetMapValueIfNotNil(workitemStatusMap, "default_destination_status", getStatusNameFromId(*workitemStatus.DefaultDestinationStatus.Id, workitemStatuses))
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
