package task_management_worktype_flow_oncreate_rule

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
)

/*
The resource_genesyscloud_task_management_oncreate_rule_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.
*/

func GenerateOnCreateRuleResource(
	resourceLabel,
	worktypeId,
	name string,
	attrs ...string,
) string {
	return fmt.Sprintf(
		`resource "genesyscloud_task_management_worktype_flow_oncreate_rule" "%s" {
		worktype_id = %s
		name = "%s"
		%s
		}
		`, resourceLabel, worktypeId, name, strings.Join(attrs, "\n"))
}

// getWorkitemoncreaterulecreateFromResourceData maps data from schema ResourceData object to a platformclientv2.Workitemoncreaterulecreate
func getWorkitemoncreaterulecreateFromResourceData(d *schema.ResourceData) platformclientv2.Workitemoncreaterulecreate {
	onCreateRule := platformclientv2.Workitemoncreaterulecreate{
		Name: platformclientv2.String(d.Get("name").(string)),
	}

	return onCreateRule
}

// getWorkitemoncreateruleupdateFromResourceData maps data from schema ResourceData object to a platformclientv2.Workitemoncreateruleupdate
func getWorkitemoncreateruleupdateFromResourceData(d *schema.ResourceData) platformclientv2.Workitemoncreateruleupdate {
	onCreateRuleUpdate := platformclientv2.Workitemoncreateruleupdate{}
	if d.HasChange("name") {
		onCreateRuleUpdate.SetField("Name", platformclientv2.String(d.Get("name").(string)))
	}
	return onCreateRuleUpdate
}

// splitWorktypeBasedTerraformId will split the rule resource id which is in the form
// <worktypeId>/<ruleId> into just the worktypeId and ruleId string
func splitWorktypeBasedTerraformId(composedId string) (worktypeId string, ruleId string) {
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
