package task_management_workitem

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v116/platformclientv2"
)

/*
The resource_genesyscloud_task_management_workitem_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.
*/

// getTaskManagementWorkitemFromResourceData maps data from schema ResourceData object to a platformclientv2.Workitem
func getTaskManagementWorkitemFromResourceData(d *schema.ResourceData) (*platformclientv2.Workitemcreate, error) {
	customFields, err := buildCustomFieldsNillable(d.Get("custom_fields").(string))
	if err != nil {
		return nil, err
	}

	return &platformclientv2.Workitemcreate{
		Name:        platformclientv2.String(d.Get("name").(string)),
		TypeId:      platformclientv2.String(d.Get("worktype_id").(string)),
		Description: platformclientv2.String(d.Get("description").(string)),
		LanguageId:  platformclientv2.String(d.Get("language_id").(string)),
		Priority:    platformclientv2.Int(d.Get("priority").(int)),

		DateDue:         resourcedata.GetNillableTime(d, "date_due"),
		DateExpires:     resourcedata.GetNillableTime(d, "expires"),
		DurationSeconds: platformclientv2.Int(d.Get("duration_seconds").(int)),
		Ttl:             platformclientv2.Int(d.Get("ttl").(int)),

		StatusId:             platformclientv2.String(d.Get("status_id").(string)),
		WorkbinId:            platformclientv2.String(d.Get("workbin_id").(string)),
		AssigneeId:           platformclientv2.String(d.Get("assignee_id").(string)),
		ExternalContactId:    platformclientv2.String(d.Get("external_contact_id").(string)),
		ExternalTag:          platformclientv2.String(d.Get("external_tag").(string)),
		QueueId:              platformclientv2.String(d.Get("queue_id").(string)),
		SkillIds:             lists.BuildSdkStringListFromInterfaceArray(d, "skills"),
		PreferredAgentIds:    lists.BuildSdkStringListFromInterfaceArray(d, "preferred_agents"),
		AutoStatusTransition: platformclientv2.Bool(d.Get("auto_status_transition").(bool)),

		CustomFields: customFields,
		ScoredAgents: buildWorkitemScoredAgents(d.Get("scored_agents").([]interface{})),
	}, nil
}

func buildCustomFieldsNillable(fieldsJson string) (*map[string]interface{}, error) {
	if fieldsJson == "" {
		return nil, nil
	}

	fieldsInterface, err := gcloud.JsonStringToInterface(fieldsJson)
	if err != nil {
		return nil, fmt.Errorf("failed to parse custom fields %s: %v", fieldsJson, err)
	}
	fieldsMap, ok := fieldsInterface.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("custom fields is not a JSON 'object': %v", fieldsJson, err)
	}

	return &fieldsMap, nil
}

// buildWorkitemScoredAgents maps an []interface{} into a Genesys Cloud *[]platformclientv2.Workitemscoredagent
func buildWorkitemScoredAgents(workitemScoredAgents []interface{}) *[]platformclientv2.Workitemscoredagentrequest {
	workitemScoredAgentsSlice := make([]platformclientv2.Workitemscoredagentrequest, 0)
	for _, workitemScoredAgent := range workitemScoredAgents {
		var sdkWorkitemScoredAgent platformclientv2.Workitemscoredagentrequest
		workitemScoredAgentsMap, ok := workitemScoredAgent.(map[string]interface{})
		if !ok {
			continue
		}

		sdkWorkitemScoredAgent.Id = platformclientv2.String(workitemScoredAgentsMap["agent_id"].(string))
		sdkWorkitemScoredAgent.Score = platformclientv2.Int(workitemScoredAgentsMap["score"].(int))

		workitemScoredAgentsSlice = append(workitemScoredAgentsSlice, sdkWorkitemScoredAgent)
	}

	return &workitemScoredAgentsSlice
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

// flattenUserReferenceWithNames maps a Genesys Cloud *[]platformclientv2.Userreferencewithname into a []interface{}
func flattenUserReferenceWithNames(userReferenceWithNames *[]platformclientv2.Userreferencewithname) []interface{} {
	if len(*userReferenceWithNames) == 0 {
		return nil
	}

	var userReferenceWithNameList []interface{}
	for _, userReferenceWithName := range *userReferenceWithNames {
		userReferenceWithNameMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(userReferenceWithNameMap, "name", userReferenceWithName.Name)

		userReferenceWithNameList = append(userReferenceWithNameList, userReferenceWithNameMap)
	}

	return userReferenceWithNameList
}

// flattenExternalContactReferences maps a Genesys Cloud *[]platformclientv2.Externalcontactreference into a []interface{}
func flattenExternalContactReferences(externalContactReferences *[]platformclientv2.Externalcontactreference) []interface{} {
	if len(*externalContactReferences) == 0 {
		return nil
	}

	var externalContactReferenceList []interface{}
	for _, externalContactReference := range *externalContactReferences {
		externalContactReferenceMap := make(map[string]interface{})

		externalContactReferenceList = append(externalContactReferenceList, externalContactReferenceMap)
	}

	return externalContactReferenceList
}

// flattenWorkitemQueueReferences maps a Genesys Cloud *[]platformclientv2.Workitemqueuereference into a []interface{}
func flattenWorkitemQueueReferences(workitemQueueReferences *[]platformclientv2.Workitemqueuereference) []interface{} {
	if len(*workitemQueueReferences) == 0 {
		return nil
	}

	var workitemQueueReferenceList []interface{}
	for _, workitemQueueReference := range *workitemQueueReferences {
		workitemQueueReferenceMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(workitemQueueReferenceMap, "name", workitemQueueReference.Name)

		workitemQueueReferenceList = append(workitemQueueReferenceList, workitemQueueReferenceMap)
	}

	return workitemQueueReferenceList
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

// flattenUserReferences maps a Genesys Cloud *[]platformclientv2.Userreference into a []interface{}
func flattenUserReferences(userReferences *[]platformclientv2.Userreference) []interface{} {
	if len(*userReferences) == 0 {
		return nil
	}

	var userReferenceList []interface{}
	for _, userReference := range *userReferences {
		userReferenceMap := make(map[string]interface{})

		userReferenceList = append(userReferenceList, userReferenceMap)
	}

	return userReferenceList
}

// flattenWorkitemScoredAgents maps a Genesys Cloud *[]platformclientv2.Workitemscoredagent into a []interface{}
func flattenWorkitemScoredAgents(workitemScoredAgents *[]platformclientv2.Workitemscoredagent) []interface{} {
	if len(*workitemScoredAgents) == 0 {
		return nil
	}

	var workitemScoredAgentList []interface{}
	for _, workitemScoredAgent := range *workitemScoredAgents {
		workitemScoredAgentMap := make(map[string]interface{})

		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(workitemScoredAgentMap, "agent", workitemScoredAgent.Agent, flattenUserReference)
		resourcedata.SetMapValueIfNotNil(workitemScoredAgentMap, "score", workitemScoredAgent.Score)

		workitemScoredAgentList = append(workitemScoredAgentList, workitemScoredAgentMap)
	}

	return workitemScoredAgentList
}
