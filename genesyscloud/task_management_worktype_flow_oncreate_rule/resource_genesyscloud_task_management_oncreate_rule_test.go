package task_management_worktype_flow_oncreate_rule

import (
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	workbin "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_workbin"
	workType "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_worktype"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

/*
The resource_genesyscloud_task_management_oncreate_rule_test.go contains all of the test cases for running the resource
tests for task_management_worktype_flow_oncreate_rule.
*/

func TestAccResourceTaskManagementOnCreateRule(t *testing.T) {
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

		// OnCreate Rule Resource
		onCreateRuleResourceLabel = "oncreate_rule_resource"
		onCreateRuleName          = "oncreate-" + uuid.NewString()
		onCreateRuleName2         = "oncreate2-" + uuid.NewString()
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
					GenerateOnCreateRuleResource(
						onCreateRuleResourceLabel,
						fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResourceLabel),
						onCreateRuleName,
						"",
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(ResourceType+"."+onCreateRuleResourceLabel, "worktype_id", fmt.Sprintf("genesyscloud_task_management_worktype.%s", wtResourceLabel), "id"),
					resource.TestCheckResourceAttr(ResourceType+"."+onCreateRuleResourceLabel, "name", onCreateRuleName),
				),
			},
			{
				// Update oncreate rule
				Config: workbin.GenerateWorkbinResource(wbResourceLabel, wbName, wbDescription, util.NullValue) +
					workType.GenerateWorktypeResourceBasic(
						wtResourceLabel,
						wtName,
						wtDescription,
						fmt.Sprintf("genesyscloud_task_management_workbin.%s.id", wbResourceLabel),
						"",
					) +
					GenerateOnCreateRuleResource(
						onCreateRuleResourceLabel,
						fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResourceLabel),
						onCreateRuleName2,
						"",
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+onCreateRuleResourceLabel, "name", onCreateRuleName2),
				),
			},
			{
				ResourceName:      ResourceType + "." + onCreateRuleResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyTaskManagementOnCreateRuleDestroyed,
	})
}

func testVerifyTaskManagementOnCreateRuleDestroyed(state *terraform.State) error {
	taskManagementApi := platformclientv2.NewTaskManagementApi()
	for _, res := range state.RootModule().Resources {
		if res.Type != ResourceType {
			continue
		}

		worktypeId, onCreateRuleId := splitWorktypeBasedTerraformId(res.Primary.ID)
		onCreateRule, resp, err := taskManagementApi.GetTaskmanagementWorktypeFlowsOncreateRule(worktypeId, onCreateRuleId)
		if onCreateRule != nil {
			return fmt.Errorf("task management oncreate rule (%s) still exists", res.Primary.ID)
		} else if util.IsStatus404(resp) {
			// Worktype no found, as expected
			continue
		} else {
			return fmt.Errorf("unexpected error: %s", err)
		}
	}

	// All oncreate rules deleted
	return nil
}
