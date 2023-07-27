package outbound_ruleset

import (
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

func getOutboundRulesetFromResourceData(d *schema.ResourceData) platformclientv2.Ruleset {
	name := d.Get("name").(string)

	return platformclientv2.Ruleset{
		Name:        &name,
		ContactList: gcloud.BuildSdkDomainEntityRef(d, "contact_list_id"),
		Queue:       gcloud.BuildSdkDomainEntityRef(d, "queue_id"),
		Rules:       buildRulesetRules(d.Get("rules").([]interface{})),
	}
}

// buildRulesetRules maps a []interface{} into a Genesys Cloud *[]platformclientv2.Dialerrule
func buildRulesetRules(rules []interface{}) *[]platformclientv2.Dialerrule {
	rulesSlice := make([]platformclientv2.Dialerrule, 0)
	for _, rule := range rules {
		var sdkRule platformclientv2.Dialerrule
		ruleMap := rule.(map[string]interface{})

		if name := ruleMap["name"].(string); name != "" {
			sdkRule.Name = &name
		}
		sdkRule.Order = platformclientv2.Int(ruleMap["order"].(int))
		if category := ruleMap["category"].(string); category != "" {
			sdkRule.Category = &category
		}
		if conditions := ruleMap["conditions"]; conditions != nil {
			sdkRule.Conditions = buildRulesetConditions(conditions.([]interface{}))
		}
		if actions := ruleMap["actions"]; actions != nil {
			sdkRule.Actions = buildRulesetActions(actions.([]interface{}))
		}

		rulesSlice = append(rulesSlice, sdkRule)
	}

	return &rulesSlice
}

// buildRulesetConditions maps a []interface{} into a Genesys Cloud *[]platformclientv2.Condition
func buildRulesetConditions(conditions []interface{}) *[]platformclientv2.Condition {
	conditionSlice := make([]platformclientv2.Condition, 0)
	for _, conditions := range conditions {
		var sdkCondition platformclientv2.Condition
		conditionMap := conditions.(map[string]interface{})

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
		if fieldMappings := conditionMap["contact_column_to_data_action_field_mappings"]; fieldMappings != nil {
			sdkCondition.ContactColumnToDataActionFieldMappings = buildRulesetContactcolumntodataactionfieldmappings(fieldMappings.([]interface{}))
		}
		if predicates := conditionMap["predicates"]; predicates != nil {
			sdkCondition.Predicates = buildPredicates(predicates.([]interface{}))
		}

		conditionSlice = append(conditionSlice, sdkCondition)
	}

	return &conditionSlice
}

// buildRulesetActions maps a []interface{} into a Genesys Cloud *[]platformclientv2.Dialeraction
func buildRulesetActions(actions []interface{}) *[]platformclientv2.Dialeraction {
	actionsSlice := make([]platformclientv2.Dialeraction, 0)
	for _, action := range actions {
		var sdkAction platformclientv2.Dialeraction
		actionMap := action.(map[string]interface{})

		if varType := actionMap["type"].(string); varType != "" {
			sdkAction.VarType = &varType
		}
		if actionTypeName := actionMap["action_type_name"].(string); actionTypeName != "" {
			sdkAction.ActionTypeName = &actionTypeName
		}
		if updateOption := actionMap["update_option"].(string); updateOption != "" {
			sdkAction.UpdateOption = &updateOption
		}
		if properties := actionMap["properties"].(map[string]interface{}); properties != nil {
			sdkProperties := map[string]string{}
			for k, v := range properties {
				sdkProperties[k] = v.(string)
			}
			sdkAction.Properties = &sdkProperties
		}
		sdkAction.DataAction = &platformclientv2.Domainentityref{Id: platformclientv2.String(actionMap["data_action_id"].(string))}
		if fieldMappings := actionMap["contact_column_to_data_action_field_mappings"]; fieldMappings != nil {
			sdkAction.ContactColumnToDataActionFieldMappings = buildRulesetContactcolumntodataactionfieldmappings(fieldMappings.([]interface{}))
		}
		if contactIdField := actionMap["contact_id_field"].(string); contactIdField != "" {
			sdkAction.ContactIdField = &contactIdField
		}
		if callAnalysisResultField := actionMap["call_analysis_result_field"].(string); callAnalysisResultField != "" {
			sdkAction.CallAnalysisResultField = &callAnalysisResultField
		}
		if agentWrapupField := actionMap["agent_wrapup_field"].(string); agentWrapupField != "" {
			sdkAction.AgentWrapupField = &agentWrapupField
		}

		actionsSlice = append(actionsSlice, sdkAction)
	}
	
	return &actionsSlice
}

// buildRulesetContactcolumntodataactionfieldmappings maps a []interface{} into a Genesys Cloud *[]platformclientv2.Contactcolumntodataactionfieldmapping
func buildRulesetContactcolumntodataactionfieldmappings(fieldmappings []interface{}) *[]platformclientv2.Contactcolumntodataactionfieldmapping {
	fieldmappingsSlice := make([]platformclientv2.Contactcolumntodataactionfieldmapping, 0)
	for _, fieldmapping := range fieldmappings {
		var sdkFieldmapping platformclientv2.Contactcolumntodataactionfieldmapping
		fieldmappingMap := fieldmapping.(map[string]interface{})

		if contactColumnName := fieldmappingMap["contact_column_name"].(string); contactColumnName != "" {
			sdkFieldmapping.ContactColumnName = &contactColumnName
		}
		if dataActionField := fieldmappingMap["data_action_field"].(string); dataActionField != "" {
			sdkFieldmapping.DataActionField = &dataActionField
		}

		fieldmappingsSlice = append(fieldmappingsSlice, sdkFieldmapping)
	}

	return &fieldmappingsSlice
}

// buildPredicates maps a []interface{} into a Genesys Cloud *[]platformclientv2.Dataactionconditionpredicate
func buildPredicates(predicates []interface{}) *[]platformclientv2.Dataactionconditionpredicate {
	predicatesSlice := make([]platformclientv2.Dataactionconditionpredicate, 0)
	for _, predicate := range predicates {
		var sdkPredicate platformclientv2.Dataactionconditionpredicate
		predicateMap := predicate.(map[string]interface{})

		if outputField := predicateMap["output_field"].(string); outputField != "" {
			sdkPredicate.OutputField = &outputField
		}
		if outputOperator := predicateMap["output_operator"].(string); outputOperator != "" {
			sdkPredicate.OutputOperator = &outputOperator
		}
		if comparisonValue := predicateMap["comparison_value"].(string); comparisonValue != "" {
			sdkPredicate.ComparisonValue = &comparisonValue
		}
		sdkPredicate.Inverted = platformclientv2.Bool(predicateMap["inverted"].(bool))
		sdkPredicate.OutputFieldMissingResolution = platformclientv2.Bool(predicateMap["output_field_missing_resolution"].(bool))

		predicatesSlice = append(predicatesSlice, sdkPredicate)
	}

	return &predicatesSlice
}

// flattenRulesetRule maps a Genesys Cloud []platformclientv2.Dialerrule into a []interface{}
func flattenRulesetRules(rules *[]platformclientv2.Dialerrule) []interface{} {
	if len(*rules) == 0 {
		return nil
	}

	var dialerruleList []interface{}
	for _, rule := range *rules {
		ruleMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(ruleMap, "name", rule.Name)
		resourcedata.SetMapValueIfNotNil(ruleMap, "order", rule.Order)
		resourcedata.SetMapValueIfNotNil(ruleMap, "category", rule.Category)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(ruleMap, "conditions", rule.Conditions, flattenRulesetRuleCondition)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(ruleMap, "actions", rule.Actions, flattenRulesetRuleAction)

		dialerruleList = append(dialerruleList, ruleMap)
	}

	return dialerruleList
}

// flattenRulesetRuleAction maps a Genesys Cloud []platformclientv2.Dialeraction into a []interface{}
func flattenRulesetRuleAction(actions *[]platformclientv2.Dialeraction) []interface{} {
	if len(*actions) == 0 {
		return nil
	}

	var actionList []interface{}
	for _, action := range *actions {
		actionMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(actionMap, "type", action.VarType)
		resourcedata.SetMapValueIfNotNil(actionMap, "action_type_name", action.ActionTypeName)
		resourcedata.SetMapValueIfNotNil(actionMap, "update_option", action.UpdateOption)
		resourcedata.SetMapValueIfNotNil(actionMap, "contact_id_field", action.ContactIdField)
		resourcedata.SetMapValueIfNotNil(actionMap, "call_analysis_result_field", action.CallAnalysisResultField)
		resourcedata.SetMapValueIfNotNil(actionMap, "agent_wrapup_field", action.AgentWrapupField)
		resourcedata.SetMapReferenceValueIfNotNil(actionMap, "data_action_id", action.DataAction)
		resourcedata.SetMapStringMapValueIfNotNil(actionMap, "properties", action.Properties)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(actionMap, "contact_column_to_data_action_field_mappings", action.ContactColumnToDataActionFieldMappings, flattenRulesetContactcolumntodataactionfieldmapping)

		actionList = append(actionList, actionMap)
	}

	return actionList
}

// flattenRulesetRuleCondition maps a Genesys Cloud []platformclientv2.Condition into a []interface{}
func flattenRulesetRuleCondition(conditions *[]platformclientv2.Condition) []interface{} {
	if len(*conditions) == 0 {
		return nil
	}

	var conditionList []interface{}
	for _, condition := range *conditions {
		conditionMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(conditionMap, "type", condition.VarType)
		resourcedata.SetMapValueIfNotNil(conditionMap, "inverted", condition.Inverted)
		resourcedata.SetMapValueIfNotNil(conditionMap, "attribute_name", condition.AttributeName)
		resourcedata.SetMapValueIfNotNil(conditionMap, "value", condition.Value)
		resourcedata.SetMapValueIfNotNil(conditionMap, "value_type", condition.ValueType)
		resourcedata.SetMapValueIfNotNil(conditionMap, "operator", condition.Operator)
		resourcedata.SetMapStringArrayValueIfNotNil(conditionMap, "codes", condition.Codes)
		resourcedata.SetMapValueIfNotNil(conditionMap, "property", condition.Property)
		resourcedata.SetMapValueIfNotNil(conditionMap, "property_type", condition.PropertyType)
		resourcedata.SetMapReferenceValueIfNotNil(conditionMap, "data_action_id", condition.DataAction)
		resourcedata.SetMapValueIfNotNil(conditionMap, "data_not_found_resolution", condition.DataNotFoundResolution)
		resourcedata.SetMapValueIfNotNil(conditionMap, "contact_id_field", condition.ContactIdField)
		resourcedata.SetMapValueIfNotNil(conditionMap, "call_analysis_result_field", condition.CallAnalysisResultField)
		resourcedata.SetMapValueIfNotNil(conditionMap, "agent_wrapup_field", condition.AgentWrapupField)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(conditionMap, "contact_column_to_data_action_field_mappings", condition.ContactColumnToDataActionFieldMappings, flattenRulesetContactcolumntodataactionfieldmapping)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(conditionMap, "predicates", condition.Predicates, flattenRulesetPredicate)

		conditionList = append(conditionList, conditionMap)
	}

	return conditionList
}

// flattenRulesetContactcolumntodataactionfieldmapping maps a Genesys Cloud []platformclientv2.Contactcolumntodataactionfieldmapping into a []interface{}
func flattenRulesetContactcolumntodataactionfieldmapping(fieldmappings *[]platformclientv2.Contactcolumntodataactionfieldmapping) []interface{} {
	if len(*fieldmappings) == 0 {
		return nil
	}

	var fieldmappingsList []interface{}
	for _, fieldmapping := range *fieldmappings {
		fieldmappingMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(fieldmappingMap, "contact_column_name", fieldmapping.ContactColumnName)
		resourcedata.SetMapValueIfNotNil(fieldmappingMap, "data_action_field", fieldmapping.DataActionField)

		fieldmappingsList = append(fieldmappingsList, fieldmappingMap)
	}

	return fieldmappingsList
}

// flattenRulesetPredicate maps a Genesys Cloud []platformclientv2.Dataactionconditionpredicate into a []interface{}
func flattenRulesetPredicate(predicates *[]platformclientv2.Dataactionconditionpredicate) []interface{} {
	if len(*predicates) == 0 {
		return nil
	}

	var predicateList []interface{}
	for _, predicate := range *predicates {
		predicateMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(predicateMap, "output_field", predicate.OutputField)
		resourcedata.SetMapValueIfNotNil(predicateMap, "output_operator", predicate.OutputOperator)
		resourcedata.SetMapValueIfNotNil(predicateMap, "comparison_value", predicate.ComparisonValue)
		resourcedata.SetMapValueIfNotNil(predicateMap, "inverted", predicate.Inverted)
		resourcedata.SetMapValueIfNotNil(predicateMap, "output_field_missing_resolution", predicate.OutputFieldMissingResolution)

		predicateList = append(predicateList, predicateMap)
	}

	return predicateList
}
