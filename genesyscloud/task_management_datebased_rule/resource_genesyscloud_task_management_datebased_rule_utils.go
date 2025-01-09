package task_management_datebased_rule

import (
	"fmt"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
)

/*
The resource_genesyscloud_task_management_datebased_rule_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.
*/

func GenerateDateBasedRuleResource(
	resourceLabel,
	worktypeId,
	name string,
	attribute string,
	relativeMinutesToInvocation int,
) string {
	return fmt.Sprintf(
		`resource "genesyscloud_task_management_datebased_rule" "%s" {
		worktype_id = %s
		name = "%s"
		condition {
			attribute = "%s"
			relative_minutes_to_invocation = %d
		}
	}
	`, resourceLabel, worktypeId, name, attribute, relativeMinutesToInvocation)
}

// getWorkitemdatebasedrulecreateFromResourceData maps data from schema ResourceData object to a platformclientv2.Workitemdatebasedrulecreate
func getWorkitemdatebasedrulecreateFromResourceData(d *schema.ResourceData) platformclientv2.Workitemdatebasedrulecreate {
	condition := d.Get("condition").([]interface{})
	conditionMap := condition[0].(map[string]interface{})
	attribute := conditionMap["attribute"].(string)
	relativeMinutesToInvocation := conditionMap["relative_minutes_to_invocation"].(int)
	
	ruleCondition := platformclientv2.Workitemdatebasedcondition{}
	ruleCondition.SetField("Attribute", platformclientv2.String(attribute))
	ruleCondition.SetField("RelativeMinutesToInvocation", platformclientv2.Int(relativeMinutesToInvocation))
	
	datebased_rule := platformclientv2.Workitemdatebasedrulecreate{
		Name: platformclientv2.String(d.Get("name").(string)),
		Condition: &ruleCondition,
	}

	return datebased_rule
}

// getWorkitemdatebasedruleupdateFromResourceData maps data from schema ResourceData object to a platformclientv2.Workitemdatebasedruleupdate
func getWorkitemdatebasedruleupdateFromResourceData(d *schema.ResourceData) platformclientv2.Workitemdatebasedruleupdate {
	condition := d.Get("condition").([]interface{})
	conditionMap := condition[0].(map[string]interface{})
	attribute := conditionMap["attribute"].(string)
	relativeMinutesToInvocation := conditionMap["relative_minutes_to_invocation"].(int)
	
	ruleCondition := platformclientv2.Workitemdatebasedconditionupdate{}
	ruleCondition.SetField("Attribute", platformclientv2.String(attribute))
	ruleCondition.SetField("RelativeMinutesToInvocation", platformclientv2.Int(relativeMinutesToInvocation))
	
	dateBasedRuleUpdate := platformclientv2.Workitemdatebasedruleupdate{}
	if d.HasChange("name") {
		dateBasedRuleUpdate.SetField("Name", platformclientv2.String(d.Get("name").(string)))
	}
	if d.HasChange("condition") {
		dateBasedRuleUpdate.SetField("Condition", &ruleCondition)
	}
	return dateBasedRuleUpdate
}

// splitWorktypeBasedTerraformId will split the rule resource id which is in the form
// <worktypeId>/<id> into just the worktypeId and id string
func splitWorktypeBasedTerraformId(composedId string) (worktypeId string, id string) {
	return strings.Split(composedId, "/")[0], strings.Split(composedId, "/")[1]
}

// flattenSdkCondition converts a *platformclientv2.Workitemdatebasedrule into a map and then into array for consumption by Terraform
func flattenSdkCondition(rule *platformclientv2.Workitemdatebasedrule) []interface{} {
	conditionInterface := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(conditionInterface, "attribute", rule.Condition.Attribute)
	resourcedata.SetMapValueIfNotNil(conditionInterface, "relative_minutes_to_invocation", rule.Condition.RelativeMinutesToInvocation)
	
	return []interface{}{conditionInterface}
}
