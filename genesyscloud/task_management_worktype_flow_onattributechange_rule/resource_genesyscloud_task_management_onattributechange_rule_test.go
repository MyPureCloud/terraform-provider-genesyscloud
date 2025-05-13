package task_management_worktype_flow_onattributechange_rule

import (
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	workbin "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_workbin"
	workType "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_worktype"
	worktypeStatus "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_worktype_status"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

/*
The resource_genesyscloud_task_management_onattributechange_rule_test.go contains all of the test cases for running the resource
tests for task_management_worktype_flow_onattributechange_rule.
*/

func TestAccResourceTaskManagementOnAttributeChangeRule(t *testing.T) {
	t.Parallel()
	var (
		// Workbin
		wbResourceLabel = "workbin_1"
		wbName          = "wb_" + uuid.NewString()
		wbDescription   = "workbin created for CX as Code test case"

		// Worktype
		wtResourceLabel = "worktype_1"
		wtName          = "wt_" + uuid.NewString()
		wtDescription   = "test worktype description"

		// Status
		statusOpenResourceLabel       = "status_1"
		statusOpenName                = "status-" + uuid.NewString()
		statusOpenCategory            = "Open"
		statusInProgressResourceLabel = "status_2"
		statusInProgressName          = "status-" + uuid.NewString()
		statusInProgressCategory      = "InProgress"
		statusClosedResourceLabel     = "status_3"
		statusClosedName              = "status-" + uuid.NewString()
		statusClosedCategory          = "Closed"

		// OnAttributeChange Rule Resource
		onAttributeChangeRuleResourceLabel = "onattributechange_rule_resource"
		onAttributeChangeRuleName          = "onattributechange-" + uuid.NewString()
		onAttributeChangeRuleName2         = "onattributechange2-" + uuid.NewString()
		onAttributeChangeRuleAttribute     = "statusId"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create worktype status
				Config: workbin.GenerateWorkbinResource(wbResourceLabel, wbName, wbDescription, util.NullValue) +
					workType.GenerateWorktypeResourceBasic(
						wtResourceLabel,
						wtName,
						wtDescription,
						fmt.Sprintf("genesyscloud_task_management_workbin.%s.id", wbResourceLabel),
						"",
					) +
					worktypeStatus.GenerateWorktypeStatusResource(
						statusInProgressResourceLabel,
						fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResourceLabel),
						statusInProgressName,
						statusInProgressCategory,
						"", util.NullValue, "",
					) +
					GenerateOnAttributeChangeRuleResource(
						onAttributeChangeRuleResourceLabel,
						fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResourceLabel),
						onAttributeChangeRuleName,
						onAttributeChangeRuleAttribute,
						fmt.Sprintf("genesyscloud_task_management_worktype_status.%s.id", statusInProgressResourceLabel),
						"", "",
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(ResourceType+"."+onAttributeChangeRuleResourceLabel, "worktype_id", fmt.Sprintf("genesyscloud_task_management_worktype.%s", wtResourceLabel), "id"),
					resource.TestCheckResourceAttr(ResourceType+"."+onAttributeChangeRuleResourceLabel, "name", onAttributeChangeRuleName),
					resource.TestCheckResourceAttrPair(ResourceType+"."+onAttributeChangeRuleResourceLabel, "condition.0.new_value",
						fmt.Sprintf("genesyscloud_task_management_worktype_status.%s", statusInProgressResourceLabel), "id"),
				),
			},
			{
				// Update onattributechange rule
				Config: workbin.GenerateWorkbinResource(wbResourceLabel, wbName, wbDescription, util.NullValue) +
					workType.GenerateWorktypeResourceBasic(
						wtResourceLabel,
						wtName,
						wtDescription,
						fmt.Sprintf("genesyscloud_task_management_workbin.%s.id", wbResourceLabel),
						"",
					) +
					worktypeStatus.GenerateWorktypeStatusResource(
						statusOpenResourceLabel,
						fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResourceLabel),
						statusOpenName,
						statusOpenCategory,
						"", util.NullValue, "",
					) +
					worktypeStatus.GenerateWorktypeStatusResource(
						statusInProgressResourceLabel,
						fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResourceLabel),
						statusInProgressName,
						statusInProgressCategory,
						"", util.NullValue, "",
					) +
					worktypeStatus.GenerateWorktypeStatusResource(
						statusClosedResourceLabel,
						fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResourceLabel),
						statusClosedName,
						statusClosedCategory,
						"", util.NullValue, "",
					) +
					GenerateOnAttributeChangeRuleResource(
						onAttributeChangeRuleResourceLabel,
						fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResourceLabel),
						onAttributeChangeRuleName2,
						onAttributeChangeRuleAttribute,
						fmt.Sprintf("genesyscloud_task_management_worktype_status.%s.id", statusClosedResourceLabel),
						fmt.Sprintf("genesyscloud_task_management_worktype_status.%s.id", statusOpenResourceLabel),
						"",
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+onAttributeChangeRuleResourceLabel, "name", onAttributeChangeRuleName2),
					resource.TestCheckResourceAttrPair(ResourceType+"."+onAttributeChangeRuleResourceLabel, "condition.0.new_value",
						fmt.Sprintf("genesyscloud_task_management_worktype_status.%s", statusClosedResourceLabel), "id"),
					resource.TestCheckResourceAttrPair(ResourceType+"."+onAttributeChangeRuleResourceLabel, "condition.0.old_value",
						fmt.Sprintf("genesyscloud_task_management_worktype_status.%s", statusOpenResourceLabel), "id"),
				),
			},
			{
				ResourceName:      ResourceType + "." + onAttributeChangeRuleResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyTaskManagementOnAttributeChangeRuleDestroyed,
	})
}

func testVerifyTaskManagementOnAttributeChangeRuleDestroyed(state *terraform.State) error {
	taskManagementApi := platformclientv2.NewTaskManagementApi()
	for _, res := range state.RootModule().Resources {
		if res.Type != ResourceType {
			continue
		}

		worktypeId, onAttributeChangeRuleId := splitWorktypeBasedTerraformId(res.Primary.ID)
		onAttributeChangeRule, resp, err := taskManagementApi.GetTaskmanagementWorktypeFlowsOnattributechangeRule(worktypeId, onAttributeChangeRuleId)
		if onAttributeChangeRule != nil {
			return fmt.Errorf("task management onattributechange rule (%s) still exists", res.Primary.ID)
		} else if util.IsStatus404(resp) {
			// Worktype no found, as expected
			continue
		} else {
			return fmt.Errorf("unexpected error: %s", err)
		}
	}

	// All onattributechange rules deleted
	return nil
}
