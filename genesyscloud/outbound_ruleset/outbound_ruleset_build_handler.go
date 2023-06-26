package outbound_ruleset

import (
	"encoding/json"

	"github.com/mypurecloud/platform-client-sdk-go/v102/platformclientv2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type ProcessAutomationTrigger struct {
	Id              *string `json:"id,omitempty"`
	TopicName       *string `json:"topicName,omitempty"`
	Name            *string `json:"name,omitempty"`
	Target          *Target `json:"target,omitempty"`
	MatchCriteria   *string `json:"-"`
	Enabled         *bool   `json:"enabled,omitempty"`
	EventTTLSeconds *int    `json:"eventTTLSeconds,omitempty"`
	DelayBySeconds  *int    `json:"delayBySeconds,omitempty"`
	Version         *int    `json:"version,omitempty"`
	Description     *string `json:"description,omitempty"`
}

type Target struct {
	Type *string `json:"type,omitempty"`
	Id   *string `json:"id,omitempty"`
}

func (p *ProcessAutomationTrigger) toJSONString() (string, error) {
	//Step #1: Converting the process automation trigger to a JSON byte arrays
	b, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	patJson := string(b)

	//Step #2: Converting the JSON string to a Golang Map
	var patMap map[string]interface{}
	err = json.Unmarshal([]byte(patJson), &patMap)
	if err != nil {
		return "", err
	}

	//Step #3: Converting the MatchCriteria field from a string to Map
	var data []map[string]interface{}
	err = json.Unmarshal([]byte(*p.MatchCriteria), &data)
	if err != nil {
		return "", err
	}

	matchCriteriaArray := make([]interface{}, len(data))
	for i, obj := range data {
		value := make(map[string]interface{})

		value["jsonPath"] = obj["jsonPath"]
		value["operator"] = obj["operator"]
		value["value"] = obj["value"]

		matchCriteriaArray[i] = value
	}

	//Step #4: Merging the match criteria array into the main map
	patMap["matchCriteria"] = matchCriteriaArray

	//Step #5: Converting the merged Map into a JSON string
	finalJsonBytes, err := json.Marshal(patMap)
	if err != nil {
		return "", err
	}

	finalPAT := string(finalJsonBytes)
	return finalPAT, nil
}

// Constructor that will take an platform client response object and build a new ProcessAutomationTrigger from it
func NewProcessAutomationFromPayload(response *platformclientv2.APIResponse) (*ProcessAutomationTrigger, error) {
	httpPayload := response.RawBody
	pat := &ProcessAutomationTrigger{}
	patMap := make(map[string]interface{})
	err := json.Unmarshal(httpPayload, &patMap)
	if err != nil {
		return nil, err
	}

	matchCriteria := patMap["matchCriteria"]
	matchCriteriaBytes, err := json.Marshal(matchCriteria)
	matchCriteriaStr := string(matchCriteriaBytes)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(httpPayload, &pat)
	if err != nil {
		return nil, err
	}
	pat.MatchCriteria = &matchCriteriaStr

	return pat, nil
}


func buildSdkoutboundrulesetContactcolumntodataactionfieldmappingSlice(contactcolumntodataactionfieldmapping *schema.Set) *[]platformclientv2.Contactcolumntodataactionfieldmapping {
	if contactcolumntodataactionfieldmapping == nil {
		return nil
	}
	sdkContactcolumntodataactionfieldmappingSlice := make([]platformclientv2.Contactcolumntodataactionfieldmapping, 0)
	contactcolumntodataactionfieldmappingList := contactcolumntodataactionfieldmapping.List()
	for _, configcontactcolumntodataactionfieldmapping := range contactcolumntodataactionfieldmappingList {
		var sdkContactcolumntodataactionfieldmapping platformclientv2.Contactcolumntodataactionfieldmapping
		contactcolumntodataactionfieldmappingMap := configcontactcolumntodataactionfieldmapping.(map[string]interface{})
		if contactColumnName := contactcolumntodataactionfieldmappingMap["contact_column_name"].(string); contactColumnName != "" {
			sdkContactcolumntodataactionfieldmapping.ContactColumnName = &contactColumnName
		}
		if dataActionField := contactcolumntodataactionfieldmappingMap["data_action_field"].(string); dataActionField != "" {
			sdkContactcolumntodataactionfieldmapping.DataActionField = &dataActionField
		}

		sdkContactcolumntodataactionfieldmappingSlice = append(sdkContactcolumntodataactionfieldmappingSlice, sdkContactcolumntodataactionfieldmapping)
	}
	return &sdkContactcolumntodataactionfieldmappingSlice
}

func buildSdkoutboundrulesetDataactionconditionpredicateSlice(dataactionconditionpredicate *schema.Set) *[]platformclientv2.Dataactionconditionpredicate {
	if dataactionconditionpredicate == nil {
		return nil
	}
	sdkDataactionconditionpredicateSlice := make([]platformclientv2.Dataactionconditionpredicate, 0)
	dataactionconditionpredicateList := dataactionconditionpredicate.List()
	for _, configdataactionconditionpredicate := range dataactionconditionpredicateList {
		var sdkDataactionconditionpredicate platformclientv2.Dataactionconditionpredicate
		dataactionconditionpredicateMap := configdataactionconditionpredicate.(map[string]interface{})
		if outputField := dataactionconditionpredicateMap["output_field"].(string); outputField != "" {
			sdkDataactionconditionpredicate.OutputField = &outputField
		}
		if outputOperator := dataactionconditionpredicateMap["output_operator"].(string); outputOperator != "" {
			sdkDataactionconditionpredicate.OutputOperator = &outputOperator
		}
		if comparisonValue := dataactionconditionpredicateMap["comparison_value"].(string); comparisonValue != "" {
			sdkDataactionconditionpredicate.ComparisonValue = &comparisonValue
		}
		sdkDataactionconditionpredicate.Inverted = platformclientv2.Bool(dataactionconditionpredicateMap["inverted"].(bool))
		sdkDataactionconditionpredicate.OutputFieldMissingResolution = platformclientv2.Bool(dataactionconditionpredicateMap["output_field_missing_resolution"].(bool))

		sdkDataactionconditionpredicateSlice = append(sdkDataactionconditionpredicateSlice, sdkDataactionconditionpredicate)
	}
	return &sdkDataactionconditionpredicateSlice
}

func buildSdkoutboundrulesetConditionSlice(conditionList []interface{}) *[]platformclientv2.Condition {
	sdkConditionSlice := make([]platformclientv2.Condition, 0)
	for _, configcondition := range conditionList {
		var sdkCondition platformclientv2.Condition
		conditionMap := configcondition.(map[string]interface{})
		if varType := conditionMap["type"].(string); varType != "" {
			sdkCondition.VarType = &varType
		}
		sdkCondition.Inverted = platformclientv2.Bool(conditionMap["inverted"].(bool))
		if attributeName := conditionMap["attribute_name"].(string); attributeName != "" {
			sdkCondition.AttributeName = &attributeName
		}
		if value := conditionMap["value"].(string); value != "" {
			sdkCondition.Value = &value
		}
		if valueType := conditionMap["value_type"].(string); valueType != "" {
			sdkCondition.ValueType = &valueType
		}
		if operator := conditionMap["operator"].(string); operator != "" {
			sdkCondition.Operator = &operator
		}
		codes := make([]string, 0)
		for _, v := range conditionMap["codes"].([]interface{}) {
			codes = append(codes, v.(string))
		}
		sdkCondition.Codes = &codes
		if property := conditionMap["property"].(string); property != "" {
			sdkCondition.Property = &property
		}
		if propertyType := conditionMap["property_type"].(string); propertyType != "" {
			sdkCondition.PropertyType = &propertyType
		}
		sdkCondition.DataAction = &platformclientv2.Domainentityref{Id: platformclientv2.String(conditionMap["data_action_id"].(string))}
		sdkCondition.DataNotFoundResolution = platformclientv2.Bool(conditionMap["data_not_found_resolution"].(bool))
		if contactIdField := conditionMap["contact_id_field"].(string); contactIdField != "" {
			sdkCondition.ContactIdField = &contactIdField
		}
		if callAnalysisResultField := conditionMap["call_analysis_result_field"].(string); callAnalysisResultField != "" {
			sdkCondition.CallAnalysisResultField = &callAnalysisResultField
		}
		if agentWrapupField := conditionMap["agent_wrapup_field"].(string); agentWrapupField != "" {
			sdkCondition.AgentWrapupField = &agentWrapupField
		}
		if contactColumnToDataActionFieldMappings := conditionMap["contact_column_to_data_action_field_mappings"]; contactColumnToDataActionFieldMappings != nil {
			sdkCondition.ContactColumnToDataActionFieldMappings = buildSdkoutboundrulesetContactcolumntodataactionfieldmappingSlice(contactColumnToDataActionFieldMappings.(*schema.Set))
		}
		if predicates := conditionMap["predicates"]; predicates != nil {
			sdkCondition.Predicates = buildSdkoutboundrulesetDataactionconditionpredicateSlice(predicates.(*schema.Set))
		}

		sdkConditionSlice = append(sdkConditionSlice, sdkCondition)
	}
	return &sdkConditionSlice
}

func buildSdkoutboundrulesetDialeractionSlice(dialeractionList []interface{}) *[]platformclientv2.Dialeraction {
	sdkDialeractionSlice := make([]platformclientv2.Dialeraction, 0)
	for _, configdialeraction := range dialeractionList {
		var sdkDialeraction platformclientv2.Dialeraction
		dialeractionMap := configdialeraction.(map[string]interface{})
		if varType := dialeractionMap["type"].(string); varType != "" {
			sdkDialeraction.VarType = &varType
		}
		if actionTypeName := dialeractionMap["action_type_name"].(string); actionTypeName != "" {
			sdkDialeraction.ActionTypeName = &actionTypeName
		}
		if updateOption := dialeractionMap["update_option"].(string); updateOption != "" {
			sdkDialeraction.UpdateOption = &updateOption
		}
		if properties := dialeractionMap["properties"].(map[string]interface{}); properties != nil {
			sdkProperties := map[string]string{}
			for k, v := range properties {
				sdkProperties[k] = v.(string)
			}
			sdkDialeraction.Properties = &sdkProperties
		}
		sdkDialeraction.DataAction = &platformclientv2.Domainentityref{Id: platformclientv2.String(dialeractionMap["data_action_id"].(string))}
		if contactColumnToDataActionFieldMappings := dialeractionMap["contact_column_to_data_action_field_mappings"]; contactColumnToDataActionFieldMappings != nil {
			sdkDialeraction.ContactColumnToDataActionFieldMappings = buildSdkoutboundrulesetContactcolumntodataactionfieldmappingSlice(contactColumnToDataActionFieldMappings.(*schema.Set))
		}
		if contactIdField := dialeractionMap["contact_id_field"].(string); contactIdField != "" {
			sdkDialeraction.ContactIdField = &contactIdField
		}
		if callAnalysisResultField := dialeractionMap["call_analysis_result_field"].(string); callAnalysisResultField != "" {
			sdkDialeraction.CallAnalysisResultField = &callAnalysisResultField
		}
		if agentWrapupField := dialeractionMap["agent_wrapup_field"].(string); agentWrapupField != "" {
			sdkDialeraction.AgentWrapupField = &agentWrapupField
		}

		sdkDialeractionSlice = append(sdkDialeractionSlice, sdkDialeraction)
	}
	return &sdkDialeractionSlice
}

func buildSdkoutboundrulesetDialerruleSlice(dialerruleList []interface{}) *[]platformclientv2.Dialerrule {
	sdkDialerruleSlice := make([]platformclientv2.Dialerrule, 0)
	for _, configdialerrule := range dialerruleList {
		var sdkDialerrule platformclientv2.Dialerrule
		dialerruleMap := configdialerrule.(map[string]interface{})
		if name := dialerruleMap["name"].(string); name != "" {
			sdkDialerrule.Name = &name
		}
		sdkDialerrule.Order = platformclientv2.Int(dialerruleMap["order"].(int))
		if category := dialerruleMap["category"].(string); category != "" {
			sdkDialerrule.Category = &category
		}
		if conditions := dialerruleMap["conditions"]; conditions != nil {
			sdkDialerrule.Conditions = buildSdkoutboundrulesetConditionSlice(conditions.([]interface{}))
		}
		if actions := dialerruleMap["actions"]; actions != nil {
			sdkDialerrule.Actions = buildSdkoutboundrulesetDialeractionSlice(actions.([]interface{}))
		}

		sdkDialerruleSlice = append(sdkDialerruleSlice, sdkDialerrule)
	}
	return &sdkDialerruleSlice
}

func flattenSdkoutboundrulesetContactcolumntodataactionfieldmappingSlice(contactcolumntodataactionfieldmappings []platformclientv2.Contactcolumntodataactionfieldmapping) *schema.Set {
	if len(contactcolumntodataactionfieldmappings) == 0 {
		return nil
	}

	contactcolumntodataactionfieldmappingSet := schema.NewSet(schema.HashResource(outboundrulesetcontactcolumntodataactionfieldmappingResource), []interface{}{})
	for _, contactcolumntodataactionfieldmapping := range contactcolumntodataactionfieldmappings {
		contactcolumntodataactionfieldmappingMap := make(map[string]interface{})

		if contactcolumntodataactionfieldmapping.ContactColumnName != nil {
			contactcolumntodataactionfieldmappingMap["contact_column_name"] = *contactcolumntodataactionfieldmapping.ContactColumnName
		}
		if contactcolumntodataactionfieldmapping.DataActionField != nil {
			contactcolumntodataactionfieldmappingMap["data_action_field"] = *contactcolumntodataactionfieldmapping.DataActionField
		}

		contactcolumntodataactionfieldmappingSet.Add(contactcolumntodataactionfieldmappingMap)
	}

	return contactcolumntodataactionfieldmappingSet
}

func flattenSdkoutboundrulesetDataactionconditionpredicateSlice(dataactionconditionpredicates []platformclientv2.Dataactionconditionpredicate) *schema.Set {
	if len(dataactionconditionpredicates) == 0 {
		return nil
	}

	dataactionconditionpredicateSet := schema.NewSet(schema.HashResource(outboundrulesetdataactionconditionpredicateResource), []interface{}{})
	for _, dataactionconditionpredicate := range dataactionconditionpredicates {
		dataactionconditionpredicateMap := make(map[string]interface{})

		if dataactionconditionpredicate.OutputField != nil {
			dataactionconditionpredicateMap["output_field"] = *dataactionconditionpredicate.OutputField
		}
		if dataactionconditionpredicate.OutputOperator != nil {
			dataactionconditionpredicateMap["output_operator"] = *dataactionconditionpredicate.OutputOperator
		}
		if dataactionconditionpredicate.ComparisonValue != nil {
			dataactionconditionpredicateMap["comparison_value"] = *dataactionconditionpredicate.ComparisonValue
		}
		if dataactionconditionpredicate.Inverted != nil {
			dataactionconditionpredicateMap["inverted"] = *dataactionconditionpredicate.Inverted
		}
		if dataactionconditionpredicate.OutputFieldMissingResolution != nil {
			dataactionconditionpredicateMap["output_field_missing_resolution"] = *dataactionconditionpredicate.OutputFieldMissingResolution
		}

		dataactionconditionpredicateSet.Add(dataactionconditionpredicateMap)
	}

	return dataactionconditionpredicateSet
}

func flattenSdkoutboundrulesetConditionSlice(conditions []platformclientv2.Condition) []interface{} {
	if len(conditions) == 0 {
		return nil
	}

	var conditionList []interface{}
	for _, condition := range conditions {
		conditionMap := make(map[string]interface{})

		if condition.VarType != nil {
			conditionMap["type"] = *condition.VarType
		}
		if condition.Inverted != nil {
			conditionMap["inverted"] = *condition.Inverted
		}
		if condition.AttributeName != nil {
			conditionMap["attribute_name"] = *condition.AttributeName
		}
		if condition.Value != nil {
			conditionMap["value"] = *condition.Value
		}
		if condition.ValueType != nil {
			conditionMap["value_type"] = *condition.ValueType
		}
		if condition.Operator != nil {
			conditionMap["operator"] = *condition.Operator
		}
		if condition.Codes != nil {
			codes := make([]string, 0)
			for _, v := range *condition.Codes {
				codes = append(codes, v)
			}
			conditionMap["codes"] = codes
		}
		if condition.Property != nil {
			conditionMap["property"] = *condition.Property
		}
		if condition.PropertyType != nil {
			conditionMap["property_type"] = *condition.PropertyType
		}
		if condition.DataAction != nil {
			conditionMap["data_action_id"] = *condition.DataAction.Id
		}
		if condition.DataNotFoundResolution != nil {
			conditionMap["data_not_found_resolution"] = *condition.DataNotFoundResolution
		}
		if condition.ContactIdField != nil {
			conditionMap["contact_id_field"] = *condition.ContactIdField
		}
		if condition.CallAnalysisResultField != nil {
			conditionMap["call_analysis_result_field"] = *condition.CallAnalysisResultField
		}
		if condition.AgentWrapupField != nil {
			conditionMap["agent_wrapup_field"] = *condition.AgentWrapupField
		}
		if condition.ContactColumnToDataActionFieldMappings != nil {
			conditionMap["contact_column_to_data_action_field_mappings"] = flattenSdkoutboundrulesetContactcolumntodataactionfieldmappingSlice(*condition.ContactColumnToDataActionFieldMappings)
		}
		if condition.Predicates != nil {
			conditionMap["predicates"] = flattenSdkoutboundrulesetDataactionconditionpredicateSlice(*condition.Predicates)
		}

		conditionList = append(conditionList, conditionMap)
	}

	return conditionList
}

func flattenSdkoutboundrulesetDialeractionSlice(dialeractions []platformclientv2.Dialeraction) []interface{} {
	if len(dialeractions) == 0 {
		return nil
	}

	var dialeractionList []interface{}
	for _, dialeraction := range dialeractions {
		dialeractionMap := make(map[string]interface{})

		if dialeraction.VarType != nil {
			dialeractionMap["type"] = *dialeraction.VarType
		}
		if dialeraction.ActionTypeName != nil {
			dialeractionMap["action_type_name"] = *dialeraction.ActionTypeName
		}
		if dialeraction.UpdateOption != nil {
			dialeractionMap["update_option"] = *dialeraction.UpdateOption
		}
		if dialeraction.Properties != nil {
			results := make(map[string]interface{})
			for k, v := range *dialeraction.Properties {
				results[k] = v
			}
			dialeractionMap["properties"] = results
		}
		if dialeraction.DataAction != nil {
			dialeractionMap["data_action_id"] = *dialeraction.DataAction.Id
		}
		if dialeraction.ContactColumnToDataActionFieldMappings != nil {
			dialeractionMap["contact_column_to_data_action_field_mappings"] = flattenSdkoutboundrulesetContactcolumntodataactionfieldmappingSlice(*dialeraction.ContactColumnToDataActionFieldMappings)
		}
		if dialeraction.ContactIdField != nil {
			dialeractionMap["contact_id_field"] = *dialeraction.ContactIdField
		}
		if dialeraction.CallAnalysisResultField != nil {
			dialeractionMap["call_analysis_result_field"] = *dialeraction.CallAnalysisResultField
		}
		if dialeraction.AgentWrapupField != nil {
			dialeractionMap["agent_wrapup_field"] = *dialeraction.AgentWrapupField
		}

		dialeractionList = append(dialeractionList, dialeractionMap)
	}

	return dialeractionList
}

func flattenSdkoutboundrulesetDialerruleSlice(dialerrules []platformclientv2.Dialerrule) []interface{} {
	if len(dialerrules) == 0 {
		return nil
	}

	var dialerruleList []interface{}
	for _, dialerrule := range dialerrules {
		dialerruleMap := make(map[string]interface{})

		if dialerrule.Name != nil {
			dialerruleMap["name"] = *dialerrule.Name
		}
		if dialerrule.Order != nil {
			dialerruleMap["order"] = *dialerrule.Order
		}
		if dialerrule.Category != nil {
			dialerruleMap["category"] = *dialerrule.Category
		}
		if dialerrule.Conditions != nil {
			dialerruleMap["conditions"] = flattenSdkoutboundrulesetConditionSlice(*dialerrule.Conditions)
		}
		if dialerrule.Actions != nil {
			dialerruleMap["actions"] = flattenSdkoutboundrulesetDialeractionSlice(*dialerrule.Actions)
		}

		dialerruleList = append(dialerruleList, dialerruleMap)
	}

	return dialerruleList
}
