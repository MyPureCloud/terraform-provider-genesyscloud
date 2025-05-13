package task_management_worktype_flow_datebased_rule

import (
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
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
		`resource "genesyscloud_task_management_worktype_flow_datebased_rule" "%s" {
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

	datebasedRule := platformclientv2.Workitemdatebasedrulecreate{
		Name:      platformclientv2.String(d.Get("name").(string)),
		Condition: &ruleCondition,
	}

	return datebasedRule
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
	if len(strings.Split(composedId, "/")) > 1 {
		return strings.Split(composedId, "/")[0], strings.Split(composedId, "/")[1]
	} else {
		log.Printf("Invalid composedId %s", composedId)
		return "", ""
	}
}

// composeWorktypeBasedTerraformId will compose the rule resource id in the form <worktypeId>/<id>
func composeWorktypeBasedTerraformId(worktypeId string, id string) (composedId string) {
	return worktypeId + "/" + id
}

// flattenSdkCondition converts a *platformclientv2.Workitemdatebasedrule into a map and then into array for consumption by Terraform
func flattenSdkCondition(rule *platformclientv2.Workitemdatebasedrule) []interface{} {
	conditionInterface := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(conditionInterface, "attribute", rule.Condition.Attribute)
	resourcedata.SetMapValueIfNotNil(conditionInterface, "relative_minutes_to_invocation", rule.Condition.RelativeMinutesToInvocation)

	return []interface{}{conditionInterface}
}

// ValidateRuleIds will check that two status ids are the same
// The id could be in the format <worktypeId>/<id>
func validateRuleIds(ruleResource1 string, key1 string, ruleResource2 string, key2 string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rule1, ok := state.RootModule().Resources[ruleResource1]
		if !ok {
			return fmt.Errorf("failed to find rule %s", ruleResource1)
		}

		rule2, ok := state.RootModule().Resources[ruleResource2]
		if !ok {
			return fmt.Errorf("failed to find rule %s", ruleResource2)
		}

		status1Id := rule1.Primary.Attributes[key1]
		if strings.Contains(status1Id, "/") {
			_, status1Id = splitWorktypeBasedTerraformId(status1Id)
		}

		status2Id := rule2.Primary.Attributes[key2]
		if strings.Contains(status2Id, "/") {
			_, status2Id = splitWorktypeBasedTerraformId(status2Id)
		}

		if status1Id != status2Id {
			attr1 := ruleResource1 + "." + key1
			attr2 := ruleResource2 + "." + key2
			return fmt.Errorf("%s not equal to %s\n %s = %s\n %s = %s", attr1, attr2, attr1, status1Id, attr2, status2Id)
		}

		return nil
	}
}
