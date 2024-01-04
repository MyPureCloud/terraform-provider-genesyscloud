package task_management_workitem

import (
	"encoding/json"
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

// getWorkitemCreateFromResourceData maps data from schema ResourceData object to a platformclientv2.Workitemcreate
func getWorkitemCreateFromResourceData(d *schema.ResourceData) (*platformclientv2.Workitemcreate, error) {
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

// getWorkitemUpdateFromResourceData maps data from schema ResourceData object to a platformclientv2.Workitemupdate
func getWorkitemUpdateFromResourceData(d *schema.ResourceData) (*platformclientv2.Workitemupdate, error) {
	customFields, err := buildCustomFieldsNillable(d.Get("custom_fields").(string))
	if err != nil {
		return nil, err
	}

	// NOTE: The only difference from  Workitemcreate is that you can't change the Worktype
	return &platformclientv2.Workitemupdate{
		Name:        platformclientv2.String(d.Get("name").(string)),
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

// buildCustomFieldsNillable builds a Genesys Cloud *[]platformclientv2.Workitemscoredagent from a JSON string
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

// flattenRoutingSkillReferences maps a Genesys Cloud *[]platformclientv2.Routingskillreference into a []interface{}
func flattenRoutingSkillReferences(routingSkillReferences *[]platformclientv2.Routingskillreference) []interface{} {
	if len(*routingSkillReferences) == 0 {
		return nil
	}

	var skillIds []interface{}
	for _, routingSkillReference := range *routingSkillReferences {
		skillIds = append(skillIds, routingSkillReference.Id)
	}

	return skillIds
}

// flattenUserReferences maps a Genesys Cloud *[]platformclientv2.Userreference into a []interface{}
func flattenUserReferences(userReferences *[]platformclientv2.Userreference) []interface{} {
	if len(*userReferences) == 0 {
		return nil
	}

	var userIds []interface{}
	for _, userReference := range *userReferences {
		userIds = append(userIds, userReference.Id)
	}

	return userIds
}

// flattenCustomFields maps a Genesys Cloud custom fields *map[string]interface{} into a JSON string
func flattenCustomFields(customFields *map[string]interface{}) (string, error) {
	if customFields == nil {
		return "", nil
	}
	cfBytes, err := json.Marshal(customFields)
	if err != nil {
		return "", fmt.Errorf("error marshalling action contract %v: %v", customFields, err)
	}
	return string(cfBytes), nil
}

// flattenWorkitemScoredAgents maps a Genesys Cloud *[]platformclientv2.Workitemscoredagent into a []interface{}
func flattenWorkitemScoredAgents(workitemScoredAgents *[]platformclientv2.Workitemscoredagent) []interface{} {
	if len(*workitemScoredAgents) == 0 {
		return nil
	}

	var workitemScoredAgentList []interface{}
	for _, workitemScoredAgent := range *workitemScoredAgents {
		workitemScoredAgentMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(workitemScoredAgentMap, "agent_id", workitemScoredAgent.Agent.Id)
		resourcedata.SetMapValueIfNotNil(workitemScoredAgentMap, "score", workitemScoredAgent.Score)

		workitemScoredAgentList = append(workitemScoredAgentList, workitemScoredAgentMap)
	}

	return workitemScoredAgentList
}
