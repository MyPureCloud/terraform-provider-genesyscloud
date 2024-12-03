package task_management_oncreate_rule

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
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
		`resource "genesyscloud_task_management_oncreate_rule" "%s" {
		worktype_id = %s
		name = "%s"
		%s
	}
`, resourceLabel, worktypeId, name, strings.Join(attrs, "\n"))
}

// getWorkitemoncreaterulecreateFromResourceData maps data from schema ResourceData object to a platformclientv2.Workitemoncreaterulecreate
func getWorkitemoncreaterulecreateFromResourceData(d *schema.ResourceData) platformclientv2.Workitemoncreaterulecreate {
	oncreate_rule := platformclientv2.Workitemoncreaterulecreate{
		Name: platformclientv2.String(d.Get("name").(string)),
	}

	return oncreate_rule
}

// getWorkitemoncreateruleupdateFromResourceData maps data from schema ResourceData object to a platformclientv2.Workitemoncreateruleupdate
func getWorkitemoncreateruleupdateFromResourceData(d *schema.ResourceData) platformclientv2.Workitemoncreateruleupdate {
	onCreateRuleUpdate := platformclientv2.Workitemoncreateruleupdate{}
	if d.HasChange("name") {
		onCreateRuleUpdate.SetField("Name", platformclientv2.String(d.Get("name").(string)))
	}
	return onCreateRuleUpdate
}

// splitOnCreateRuleTerraformId will split the status resource id which is in the form
// <worktypeId>/<ruleId> into just the worktypeId and ruleId string
func splitOnCreateRuleTerraformId(id string) (worktypeId string, ruleId string) {
	return strings.Split(id, "/")[0], strings.Split(id, "/")[1]
}
