package outbound_ruleset

import (
	"encoding/json"
	"log"
	"strings"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The resource_genesyscloud_outbound_ruleset_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.
*/

// getOutboundRulesetFromResourceData maps data from schema ResourceData object to a platformclientv2.Ruleset
func getOutboundRulesetFromResourceData(d *schema.ResourceData) platformclientv2.Ruleset {
	name := d.Get("name").(string)

	return platformclientv2.Ruleset{
		Name:        &name,
		ContactList: util.BuildSdkDomainEntityRef(d, "contact_list_id"),
		Queue:       util.BuildSdkDomainEntityRef(d, "queue_id"),
		Rules:       buildDialerules(d.Get("rules").([]interface{})),
	}
}

// buildDialerules maps a []interface{} into a Genesys Cloud *[]platformclientv2.Dialerrule
func buildDialerules(rules []interface{}) *[]platformclientv2.Dialerrule {
	rulesSlice := make([]platformclientv2.Dialerrule, 0)
	for _, rule := range rules {
		var sdkRule platformclientv2.Dialerrule
		ruleMap, ok := rule.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkRule.Name, ruleMap, "name")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkRule.Category, ruleMap, "category")
		sdkRule.Order = platformclientv2.Int(ruleMap["order"].(int))
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkRule.Conditions, ruleMap, "conditions", buildConditions)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkRule.Actions, ruleMap, "actions", buildDialeractions)

		rulesSlice = append(rulesSlice, sdkRule)
	}

	return &rulesSlice
}

// buildConditions maps a []interface{} into a Genesys Cloud *[]platformclientv2.Condition
func buildConditions(conditions []interface{}) *[]platformclientv2.Condition {
	conditionSlice := make([]platformclientv2.Condition, 0)
	for _, conditions := range conditions {
		var sdkCondition platformclientv2.Condition
		conditionMap, ok := conditions.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkCondition.VarType, conditionMap, "type")
		sdkCondition.Inverted = platformclientv2.Bool(conditionMap["inverted"].(bool))
		resourcedata.BuildSDKStringValueIfNotNil(&sdkCondition.AttributeName, conditionMap, "attribute_name")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkCondition.Value, conditionMap, "value")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkCondition.ValueType, conditionMap, "value_type")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkCondition.Operator, conditionMap, "operator")
		resourcedata.BuildSDKStringArrayValueIfNotNil(&sdkCondition.Codes, conditionMap, "codes")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkCondition.Property, conditionMap, "property")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkCondition.PropertyType, conditionMap, "property_type")
		sdkCondition.DataAction = &platformclientv2.Domainentityref{Id: platformclientv2.String(conditionMap["data_action_id"].(string))}
		sdkCondition.DataNotFoundResolution = platformclientv2.Bool(conditionMap["data_not_found_resolution"].(bool))
		resourcedata.BuildSDKStringValueIfNotNil(&sdkCondition.ContactIdField, conditionMap, "contact_id_field")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkCondition.CallAnalysisResultField, conditionMap, "call_analysis_result_field")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkCondition.AgentWrapupField, conditionMap, "agent_wrapup_field")
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkCondition.ContactColumnToDataActionFieldMappings, conditionMap, "contact_column_to_data_action_field_mappings", buildContactcolumntodataactionfieldmappings)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkCondition.Predicates, conditionMap, "predicates", buildDataactionconditionpredicates)

		conditionSlice = append(conditionSlice, sdkCondition)
	}

	return &conditionSlice
}

// buildDialeractions maps a []interface{} into a Genesys Cloud *[]platformclientv2.Dialeraction
func buildDialeractions(actions []interface{}) *[]platformclientv2.Dialeraction {
	actionsSlice := make([]platformclientv2.Dialeraction, 0)
	for _, action := range actions {
		var sdkAction platformclientv2.Dialeraction
		actionMap, ok := action.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkAction.VarType, actionMap, "type")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkAction.ActionTypeName, actionMap, "action_type_name")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkAction.UpdateOption, actionMap, "update_option")
		resourcedata.BuildSDKStringMapValueIfNotNil(&sdkAction.Properties, actionMap, "properties")
		sdkAction.DataAction = &platformclientv2.Domainentityref{Id: platformclientv2.String(actionMap["data_action_id"].(string))}
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkAction.ContactColumnToDataActionFieldMappings, actionMap, "contact_column_to_data_action_field_mappings", buildContactcolumntodataactionfieldmappings)
		resourcedata.BuildSDKStringValueIfNotNil(&sdkAction.ContactIdField, actionMap, "contact_id_field")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkAction.CallAnalysisResultField, actionMap, "call_analysis_result_field")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkAction.AgentWrapupField, actionMap, "agent_wrapup_field")

		actionsSlice = append(actionsSlice, sdkAction)
	}

	return &actionsSlice
}

// buildContactcolumntodataactionfieldmappings maps a []interface{} into a Genesys Cloud *[]platformclientv2.Contactcolumntodataactionfieldmapping
func buildContactcolumntodataactionfieldmappings(fieldmappings []interface{}) *[]platformclientv2.Contactcolumntodataactionfieldmapping {
	fieldmappingsSlice := make([]platformclientv2.Contactcolumntodataactionfieldmapping, 0)
	for _, fieldmapping := range fieldmappings {
		var sdkFieldmapping platformclientv2.Contactcolumntodataactionfieldmapping
		fieldmappingMap, ok := fieldmapping.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkFieldmapping.ContactColumnName, fieldmappingMap, "contact_column_name")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkFieldmapping.DataActionField, fieldmappingMap, "data_action_field")

		fieldmappingsSlice = append(fieldmappingsSlice, sdkFieldmapping)
	}

	return &fieldmappingsSlice
}

// buildDataactionconditionpredicates maps a []interface{} into a Genesys Cloud *[]platformclientv2.Dataactionconditionpredicate
func buildDataactionconditionpredicates(predicates []interface{}) *[]platformclientv2.Dataactionconditionpredicate {
	predicatesSlice := make([]platformclientv2.Dataactionconditionpredicate, 0)
	for _, predicate := range predicates {
		var sdkPredicate platformclientv2.Dataactionconditionpredicate
		predicateMap, ok := predicate.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkPredicate.OutputField, predicateMap, "output_field")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkPredicate.OutputOperator, predicateMap, "output_operator")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkPredicate.ComparisonValue, predicateMap, "comparison_value")
		sdkPredicate.Inverted = platformclientv2.Bool(predicateMap["inverted"].(bool))
		sdkPredicate.OutputFieldMissingResolution = platformclientv2.Bool(predicateMap["output_field_missing_resolution"].(bool))

		predicatesSlice = append(predicatesSlice, sdkPredicate)
	}

	return &predicatesSlice
}

// flattenDialerrules maps a Genesys Cloud *[]platformclientv2.Dialerrule into a []interface{}
func flattenDialerrules(rules *[]platformclientv2.Dialerrule) []interface{} {
	if len(*rules) == 0 {
		return nil
	}

	var dialerruleList []interface{}
	for _, rule := range *rules {
		ruleMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(ruleMap, "name", rule.Name)
		resourcedata.SetMapValueIfNotNil(ruleMap, "order", rule.Order)
		resourcedata.SetMapValueIfNotNil(ruleMap, "category", rule.Category)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(ruleMap, "conditions", rule.Conditions, flattenConditions)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(ruleMap, "actions", rule.Actions, flattenDialeractions)

		dialerruleList = append(dialerruleList, ruleMap)
	}

	return dialerruleList
}

// flattenDialeractions maps a Genesys Cloud *[]platformclientv2.Dialeraction into a []interface{}
func flattenDialeractions(actions *[]platformclientv2.Dialeraction) []interface{} {
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
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(actionMap, "contact_column_to_data_action_field_mappings", action.ContactColumnToDataActionFieldMappings, flattenContactcolumntodataactionfieldmappings)

		actionList = append(actionList, actionMap)
	}

	return actionList
}

// flattenConditions maps a Genesys Cloud *[]platformclientv2.Condition into a []interface{}
func flattenConditions(conditions *[]platformclientv2.Condition) []interface{} {
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
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(conditionMap, "contact_column_to_data_action_field_mappings", condition.ContactColumnToDataActionFieldMappings, flattenContactcolumntodataactionfieldmappings)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(conditionMap, "predicates", condition.Predicates, flattenDataactionconditionpredicates)

		conditionList = append(conditionList, conditionMap)
	}
	return conditionList
}

// flattenContactcolumntodataactionfieldmappings maps a Genesys Cloud *[]platformclientv2.Contactcolumntodataactionfieldmapping into a []interface{}
func flattenContactcolumntodataactionfieldmappings(fieldmappings *[]platformclientv2.Contactcolumntodataactionfieldmapping) []interface{} {
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

// flattenDataactionconditionpredicates maps a Genesys Cloud *[]platformclientv2.Dataactionconditionpredicate into a []interface{}
func flattenDataactionconditionpredicates(predicates *[]platformclientv2.Dataactionconditionpredicate) []interface{} {
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

// look through rule actions to check if the referenced skills exist in our skill map or not
func doesRuleActionsRefDeletedSkill(rule platformclientv2.Dialerrule, skillMap resourceExporter.ResourceIDMetaMap) bool {
	if rule.Actions == nil {
		return false
	}
	for _, action := range *rule.Actions {
		if action.ActionTypeName != nil && strings.EqualFold(*action.ActionTypeName, "set_skills") && action.Properties != nil {
			if value, found := (*action.Properties)["skills"]; found {
				// the property value is a json string wrapping an array of skill ids, need to convert it back to a slice to check if each skill exists
				var skillIds []string
				err := json.Unmarshal([]byte(value), &skillIds)
				if err != nil {
					log.Printf("doesRuleActionsRefDeletedSkill.Error unmarshaling skills JSON: %s", err)
					return true
				}
				for _, skillId := range skillIds {
					_, found := skillMap[skillId]
					if !found { // skill id referenced by the rule action is not found in the skill map
						log.Printf("The skill id '%s' used in action does not exist in GC anymore", skillId)
						return true
					}
				}
			}
		}
	}
	return false
}

// look through rule conditions to check if the referenced skills exist in our skill map or not
func doesRuleConditionsRefDeletedSkill(rule platformclientv2.Dialerrule, skillMap resourceExporter.ResourceIDMetaMap) bool {
	for _, condition := range *rule.Conditions {
		if condition.AttributeName != nil && strings.EqualFold(*condition.AttributeName, "skill") && condition.Value != nil {
			var found bool
			for _, value := range skillMap {
				if value.Name == *condition.Value {
					found = true
					break // found skill, evaluate next condition
				}
			}
			if !found { // skill name referenced by rule condition is not found in the skill map
				log.Printf("The skill name '%s' used in condition does not exist in GC anymore", *condition.Value)
				return true
			}
		}
	}
	return false
}
