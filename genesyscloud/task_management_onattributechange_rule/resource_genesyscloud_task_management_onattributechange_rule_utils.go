package task_management_onattributechange_rule

import (
	"fmt"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

/*
The resource_genesyscloud_task_management_onattributechange_rule_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.
*/

func GenerateOnAttributeChangeRuleResource(
	resourceLabel,
	worktypeId,
	name string,
	attribute string,
	newValue string,
	oldValue string,
	attrs ...string,
) string {
	oldValueCondition := ""
	if oldValue != "" {
		oldValueCondition = "old_value = " + oldValue
	}
	return fmt.Sprintf(
		`resource "genesyscloud_task_management_onattributechange_rule" "%s" {
		worktype_id = %s
		name = "%s"
		condition {
			attribute = "%s"
			new_value = %s
			%s
		}
		%s
	}
	`, resourceLabel, worktypeId, name, attribute, newValue, oldValueCondition, strings.Join(attrs, "\n"))
}

// getWorkitemonattributechangerulecreateFromResourceData maps data from schema ResourceData object to a platformclientv2.Workitemonattributechangerulecreate
func getWorkitemonattributechangerulecreateFromResourceData(d *schema.ResourceData) platformclientv2.Workitemonattributechangerulecreate {
	condition := d.Get("condition").([]interface{})
	conditionMap := condition[0].(map[string]interface{})
	attribute := conditionMap["attribute"].(string)
	newValue := conditionMap["new_value"].(string)
	oldValue := conditionMap["old_value"].(string)
	if attribute == "statusId" {
		_, newValue = splitWorktypeBasedTerraformId(newValue)
		if oldValue != "" {
			_, oldValue = splitWorktypeBasedTerraformId(oldValue)
		}
	}

	ruleCondition := platformclientv2.Workitemonattributechangecondition{}
	ruleCondition.SetField("Attribute", platformclientv2.String(attribute))
	ruleCondition.SetField("NewValue", platformclientv2.String(newValue))
	if oldValue != "" {
		ruleCondition.SetField("OldValue", platformclientv2.String(oldValue))
	}

	onattributechange_rule := platformclientv2.Workitemonattributechangerulecreate{
		Name: platformclientv2.String(d.Get("name").(string)),
		Condition: &ruleCondition,
	}

	return onattributechange_rule
}

// getWorkitemonattributechangeruleupdateFromResourceData maps data from schema ResourceData object to a platformclientv2.Workitemonattributechangeruleupdate
func getWorkitemonattributechangeruleupdateFromResourceData(d *schema.ResourceData) platformclientv2.Workitemonattributechangeruleupdate {
	condition := d.Get("condition").([]interface{})
	conditionMap := condition[0].(map[string]interface{})
	attribute := conditionMap["attribute"].(string)
	newValue := conditionMap["new_value"].(string)
	oldValue := conditionMap["old_value"].(string)
	if attribute == "statusId" {
		_, newValue = splitWorktypeBasedTerraformId(newValue)
		if oldValue != "" {
			_, oldValue = splitWorktypeBasedTerraformId(oldValue)
		}
	}

	ruleCondition := platformclientv2.Workitemonattributechangeconditionupdate{}
	ruleCondition.SetField("Attribute", platformclientv2.String(attribute))
	ruleCondition.SetField("NewValue", platformclientv2.String(newValue))
	if oldValue != "" {
		ruleCondition.SetField("OldValue", platformclientv2.String(oldValue))
	} else {
		ruleCondition.SetField("OldValue", nil)
	}
	
	onCreateRuleUpdate := platformclientv2.Workitemonattributechangeruleupdate{}
	if d.HasChange("name") {
		onCreateRuleUpdate.SetField("Name", platformclientv2.String(d.Get("name").(string)))
	}
	if d.HasChange("condition") {
		onCreateRuleUpdate.SetField("Condition", &ruleCondition)
	}
	return onCreateRuleUpdate
}

// splitWorktypeBasedTerraformId will split the rule resource id which is in the form
// <worktypeId>/<id> into just the worktypeId and id string
func splitWorktypeBasedTerraformId(composedId string) (worktypeId string, id string) {
	return strings.Split(composedId, "/")[0], strings.Split(composedId, "/")[1]
}

// flattenSdkCondition converts a *platformclientv2.Workitemonattributechangerule into a map and then into array for consumption by Terraform
func flattenSdkCondition(rule *platformclientv2.Workitemonattributechangerule) []interface{} {
	conditionInterface := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(conditionInterface, "attribute", rule.Condition.Attribute)
	
	if *rule.Condition.Attribute == "statusId" {
		newValue := *rule.Worktype.Id + "/" + *rule.Condition.NewValue
		resourcedata.SetMapValueIfNotNil(conditionInterface, "new_value", &newValue)

		if rule.Condition.OldValue != nil {
			oldValue := *rule.Worktype.Id + "/" + *rule.Condition.OldValue
			resourcedata.SetMapValueIfNotNil(conditionInterface, "old_value", &oldValue)
		}
	} else {
		resourcedata.SetMapValueIfNotNil(conditionInterface, "new_value", rule.Condition.NewValue)
		resourcedata.SetMapValueIfNotNil(conditionInterface, "old_value", rule.Condition.OldValue)
	}
	
	return []interface{}{conditionInterface}
}
