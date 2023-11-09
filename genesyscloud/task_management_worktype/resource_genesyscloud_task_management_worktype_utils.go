package task_management_worktype

import (
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

/*
The resource_genesyscloud_task_management_worktype_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.
*/

// getTaskManagementWorktypeFromResourceData maps data from schema ResourceData object to a platformclientv2.Worktype
func getTaskManagementWorktypeFromResourceData(d *schema.ResourceData) platformclientv2.Worktype {
	return platformclientv2.Worktype{
		Name:                      platformclientv2.String(d.Get("name").(string)),
		Division:                  &platformclientv2.Writabledivision{Id: platformclientv2.String(d.Get("division_id").(string))},
		Description:               platformclientv2.String(d.Get("description").(string)),
		DefaultWorkbin:            buildWorkbinReference(d.Get("default_workbin").([]interface{})),
		DefaultStatus:             buildWorkitemStatusReference(d.Get("default_status").([]interface{})),
		Statuses:                  buildWorkitemStatuss(d.Get("statuses").([]interface{})),
		DefaultDurationSeconds:    platformclientv2.Int(d.Get("default_duration_seconds").(int)),
		DefaultExpirationSeconds:  platformclientv2.Int(d.Get("default_expiration_seconds").(int)),
		DefaultDueDurationSeconds: platformclientv2.Int(d.Get("default_due_duration_seconds").(int)),
		DefaultPriority:           platformclientv2.Int(d.Get("default_priority").(int)),
		DefaultLanguage:           buildLanguageReference(d.Get("default_language").([]interface{})),
		DefaultTtlSeconds:         platformclientv2.Int(d.Get("default_ttl_seconds").(int)),
		DefaultQueue:              buildQueueReference(d.Get("default_queue").([]interface{})),
		DefaultSkills:             buildRoutingSkillReferences(d.Get("default_skills").([]interface{})),
		AssignmentEnabled:         platformclientv2.Bool(d.Get("assignment_enabled").(bool)),
		Schema:                    buildWorkitemSchema(d.Get("schema").([]interface{})),
	}
}

// buildWorkbinReferences maps an []interface{} into a Genesys Cloud *[]platformclientv2.Workbinreference
func buildWorkbinReferences(workbinReferences []interface{}) *[]platformclientv2.Workbinreference {
	workbinReferencesSlice := make([]platformclientv2.Workbinreference, 0)
	for _, workbinReference := range workbinReferences {
		var sdkWorkbinReference platformclientv2.Workbinreference
		workbinReferencesMap, ok := workbinReference.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkWorkbinReference.Name, workbinReferencesMap, "name")

		workbinReferencesSlice = append(workbinReferencesSlice, sdkWorkbinReference)
	}

	return &workbinReferencesSlice
}

// buildWorkitemStatusReferences maps an []interface{} into a Genesys Cloud *[]platformclientv2.Workitemstatusreference
func buildWorkitemStatusReferences(workitemStatusReferences []interface{}) *[]platformclientv2.Workitemstatusreference {
	workitemStatusReferencesSlice := make([]platformclientv2.Workitemstatusreference, 0)
	for _, workitemStatusReference := range workitemStatusReferences {
		var sdkWorkitemStatusReference platformclientv2.Workitemstatusreference
		workitemStatusReferencesMap, ok := workitemStatusReference.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkWorkitemStatusReference.Name, workitemStatusReferencesMap, "name")

		workitemStatusReferencesSlice = append(workitemStatusReferencesSlice, sdkWorkitemStatusReference)
	}

	return &workitemStatusReferencesSlice
}

// buildLocalTimes maps an []interface{} into a Genesys Cloud *[]platformclientv2.Localtime
func buildLocalTimes(localTimes []interface{}) *[]platformclientv2.Localtime {
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
		sdkLocalTime.Nano = platformclientv2.Int(localTimesMap["nano"].(int))

		localTimesSlice = append(localTimesSlice, sdkLocalTime)
	}

	return &localTimesSlice
}

// buildWorktypeReferences maps an []interface{} into a Genesys Cloud *[]platformclientv2.Worktypereference
func buildWorktypeReferences(worktypeReferences []interface{}) *[]platformclientv2.Worktypereference {
	worktypeReferencesSlice := make([]platformclientv2.Worktypereference, 0)
	for _, worktypeReference := range worktypeReferences {
		var sdkWorktypeReference platformclientv2.Worktypereference
		worktypeReferencesMap, ok := worktypeReference.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkWorktypeReference.Name, worktypeReferencesMap, "name")

		worktypeReferencesSlice = append(worktypeReferencesSlice, sdkWorktypeReference)
	}

	return &worktypeReferencesSlice
}

// buildWorkitemStatuss maps an []interface{} into a Genesys Cloud *[]platformclientv2.Workitemstatus
func buildWorkitemStatuss(workitemStatuss []interface{}) *[]platformclientv2.Workitemstatus {
	workitemStatussSlice := make([]platformclientv2.Workitemstatus, 0)
	for _, workitemStatus := range workitemStatuss {
		var sdkWorkitemStatus platformclientv2.Workitemstatus
		workitemStatussMap, ok := workitemStatus.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkWorkitemStatus.Name, workitemStatussMap, "name")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkWorkitemStatus.Category, workitemStatussMap, "category")
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkWorkitemStatus.DestinationStatuses, workitemStatussMap, "destination_statuses", buildWorkitemStatusReferences)
		resourcedata.BuildSDKStringValueIfNotNil(&sdkWorkitemStatus.Description, workitemStatussMap, "description")
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkWorkitemStatus.DefaultDestinationStatus, workitemStatussMap, "default_destination_status", buildWorkitemStatusReference)
		sdkWorkitemStatus.StatusTransitionDelaySeconds = platformclientv2.Int(workitemStatussMap["status_transition_delay_seconds"].(int))
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkWorkitemStatus.StatusTransitionTime, workitemStatussMap, "status_transition_time", buildLocalTime)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkWorkitemStatus.Worktype, workitemStatussMap, "worktype", buildWorktypeReference)

		workitemStatussSlice = append(workitemStatussSlice, sdkWorkitemStatus)
	}

	return &workitemStatussSlice
}

// buildLanguageReferences maps an []interface{} into a Genesys Cloud *[]platformclientv2.Languagereference
func buildLanguageReferences(languageReferences []interface{}) *[]platformclientv2.Languagereference {
	languageReferencesSlice := make([]platformclientv2.Languagereference, 0)
	for _, languageReference := range languageReferences {
		var sdkLanguageReference platformclientv2.Languagereference
		languageReferencesMap, ok := languageReference.(map[string]interface{})
		if !ok {
			continue
		}

		languageReferencesSlice = append(languageReferencesSlice, sdkLanguageReference)
	}

	return &languageReferencesSlice
}

// buildQueueReferences maps an []interface{} into a Genesys Cloud *[]platformclientv2.Queuereference
func buildQueueReferences(queueReferences []interface{}) *[]platformclientv2.Queuereference {
	queueReferencesSlice := make([]platformclientv2.Queuereference, 0)
	for _, queueReference := range queueReferences {
		var sdkQueueReference platformclientv2.Queuereference
		queueReferencesMap, ok := queueReference.(map[string]interface{})
		if !ok {
			continue
		}

		queueReferencesSlice = append(queueReferencesSlice, sdkQueueReference)
	}

	return &queueReferencesSlice
}

// buildRoutingSkillReferences maps an []interface{} into a Genesys Cloud *[]platformclientv2.Routingskillreference
func buildRoutingSkillReferences(routingSkillReferences []interface{}) *[]platformclientv2.Routingskillreference {
	routingSkillReferencesSlice := make([]platformclientv2.Routingskillreference, 0)
	for _, routingSkillReference := range routingSkillReferences {
		var sdkRoutingSkillReference platformclientv2.Routingskillreference
		routingSkillReferencesMap, ok := routingSkillReference.(map[string]interface{})
		if !ok {
			continue
		}

		routingSkillReferencesSlice = append(routingSkillReferencesSlice, sdkRoutingSkillReference)
	}

	return &routingSkillReferencesSlice
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

// flattenWorkbinReferences maps a Genesys Cloud *[]platformclientv2.Workbinreference into a []interface{}
func flattenWorkbinReferences(workbinReferences *[]platformclientv2.Workbinreference) []interface{} {
	if len(*workbinReferences) == 0 {
		return nil
	}

	var workbinReferenceList []interface{}
	for _, workbinReference := range *workbinReferences {
		workbinReferenceMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(workbinReferenceMap, "name", workbinReference.Name)

		workbinReferenceList = append(workbinReferenceList, workbinReferenceMap)
	}

	return workbinReferenceList
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

// flattenLocalTimes maps a Genesys Cloud *[]platformclientv2.Localtime into a []interface{}
func flattenLocalTimes(localTimes *[]platformclientv2.Localtime) []interface{} {
	if len(*localTimes) == 0 {
		return nil
	}

	var localTimeList []interface{}
	for _, localTime := range *localTimes {
		localTimeMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(localTimeMap, "hour", localTime.Hour)
		resourcedata.SetMapValueIfNotNil(localTimeMap, "minute", localTime.Minute)
		resourcedata.SetMapValueIfNotNil(localTimeMap, "second", localTime.Second)
		resourcedata.SetMapValueIfNotNil(localTimeMap, "nano", localTime.Nano)

		localTimeList = append(localTimeList, localTimeMap)
	}

	return localTimeList
}

// flattenWorktypeReferences maps a Genesys Cloud *[]platformclientv2.Worktypereference into a []interface{}
func flattenWorktypeReferences(worktypeReferences *[]platformclientv2.Worktypereference) []interface{} {
	if len(*worktypeReferences) == 0 {
		return nil
	}

	var worktypeReferenceList []interface{}
	for _, worktypeReference := range *worktypeReferences {
		worktypeReferenceMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(worktypeReferenceMap, "name", worktypeReference.Name)

		worktypeReferenceList = append(worktypeReferenceList, worktypeReferenceMap)
	}

	return worktypeReferenceList
}

// flattenWorkitemStatuss maps a Genesys Cloud *[]platformclientv2.Workitemstatus into a []interface{}
func flattenWorkitemStatuss(workitemStatuss *[]platformclientv2.Workitemstatus) []interface{} {
	if len(*workitemStatuss) == 0 {
		return nil
	}

	var workitemStatusList []interface{}
	for _, workitemStatus := range *workitemStatuss {
		workitemStatusMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(workitemStatusMap, "name", workitemStatus.Name)
		resourcedata.SetMapValueIfNotNil(workitemStatusMap, "category", workitemStatus.Category)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(workitemStatusMap, "destination_statuses", workitemStatus.DestinationStatuses, flattenWorkitemStatusReferences)
		resourcedata.SetMapValueIfNotNil(workitemStatusMap, "description", workitemStatus.Description)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(workitemStatusMap, "default_destination_status", workitemStatus.DefaultDestinationStatus, flattenWorkitemStatusReference)
		resourcedata.SetMapValueIfNotNil(workitemStatusMap, "status_transition_delay_seconds", workitemStatus.StatusTransitionDelaySeconds)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(workitemStatusMap, "status_transition_time", workitemStatus.StatusTransitionTime, flattenLocalTime)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(workitemStatusMap, "worktype", workitemStatus.Worktype, flattenWorktypeReference)

		workitemStatusList = append(workitemStatusList, workitemStatusMap)
	}

	return workitemStatusList
}

// flattenLanguageReferences maps a Genesys Cloud *[]platformclientv2.Languagereference into a []interface{}
func flattenLanguageReferences(languageReferences *[]platformclientv2.Languagereference) []interface{} {
	if len(*languageReferences) == 0 {
		return nil
	}

	var languageReferenceList []interface{}
	for _, languageReference := range *languageReferences {
		languageReferenceMap := make(map[string]interface{})

		languageReferenceList = append(languageReferenceList, languageReferenceMap)
	}

	return languageReferenceList
}

// flattenQueueReferences maps a Genesys Cloud *[]platformclientv2.Queuereference into a []interface{}
func flattenQueueReferences(queueReferences *[]platformclientv2.Queuereference) []interface{} {
	if len(*queueReferences) == 0 {
		return nil
	}

	var queueReferenceList []interface{}
	for _, queueReference := range *queueReferences {
		queueReferenceMap := make(map[string]interface{})

		queueReferenceList = append(queueReferenceList, queueReferenceMap)
	}

	return queueReferenceList
}

// flattenRoutingSkillReferences maps a Genesys Cloud *[]platformclientv2.Routingskillreference into a []interface{}
func flattenRoutingSkillReferences(routingSkillReferences *[]platformclientv2.Routingskillreference) []interface{} {
	if len(*routingSkillReferences) == 0 {
		return nil
	}

	var routingSkillReferenceList []interface{}
	for _, routingSkillReference := range *routingSkillReferences {
		routingSkillReferenceMap := make(map[string]interface{})

		routingSkillReferenceList = append(routingSkillReferenceList, routingSkillReferenceMap)
	}

	return routingSkillReferenceList
}

// flattenWorkitemSchemas maps a Genesys Cloud *[]platformclientv2.Workitemschema into a []interface{}
func flattenWorkitemSchemas(workitemSchemas *[]platformclientv2.Workitemschema) []interface{} {
	if len(*workitemSchemas) == 0 {
		return nil
	}

	var workitemSchemaList []interface{}
	for _, workitemSchema := range *workitemSchemas {
		workitemSchemaMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(workitemSchemaMap, "name", workitemSchema.Name)

		workitemSchemaList = append(workitemSchemaList, workitemSchemaMap)
	}

	return workitemSchemaList
}
