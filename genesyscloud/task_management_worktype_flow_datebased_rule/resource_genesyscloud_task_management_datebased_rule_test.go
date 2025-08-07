package task_management_worktype_flow_datebased_rule

import (
	"fmt"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	workbin "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_workbin"
	workType "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_worktype"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

/*
The resource_genesyscloud_task_management_datebased_rule_test.go contains all of the test cases for running the resource
tests for task_management_worktype_flow_datebased_rule.
*/

func TestAccResourceTaskManagementDateBasedRule(t *testing.T) {
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

		// DateBased Rule Resource
		dateBasedRuleResourceLabel   = "datebased_rule_resource"
		dateBasedRuleName            = "datebased-" + uuid.NewString()
		dateBasedRuleName2           = "datebased2-" + uuid.NewString()
		dateBasedRuleAttribute       = "dateDue"
		relativeMinutesToInvocation  = 30
		relativeMinutesToInvocation2 = 60
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
					GenerateDateBasedRuleResource(
						dateBasedRuleResourceLabel,
						fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResourceLabel),
						dateBasedRuleName,
						dateBasedRuleAttribute,
						relativeMinutesToInvocation,
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(ResourceType+"."+dateBasedRuleResourceLabel, "worktype_id", fmt.Sprintf("genesyscloud_task_management_worktype.%s", wtResourceLabel), "id"),
					resource.TestCheckResourceAttr(ResourceType+"."+dateBasedRuleResourceLabel, "name", dateBasedRuleName),
					resource.TestCheckResourceAttr(ResourceType+"."+dateBasedRuleResourceLabel, "condition.0.attribute", dateBasedRuleAttribute),
					resource.TestCheckResourceAttr(ResourceType+"."+dateBasedRuleResourceLabel, "condition.0.relative_minutes_to_invocation", fmt.Sprintf("%d", relativeMinutesToInvocation)),
				),
			},
			{
				// Update datebased rule
				Config: workbin.GenerateWorkbinResource(wbResourceLabel, wbName, wbDescription, util.NullValue) +
					workType.GenerateWorktypeResourceBasic(
						wtResourceLabel,
						wtName,
						wtDescription,
						fmt.Sprintf("genesyscloud_task_management_workbin.%s.id", wbResourceLabel),
						"",
					) +
					GenerateDateBasedRuleResource(
						dateBasedRuleResourceLabel,
						fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResourceLabel),
						dateBasedRuleName2,
						dateBasedRuleAttribute,
						relativeMinutesToInvocation2,
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+dateBasedRuleResourceLabel, "name", dateBasedRuleName2),
					resource.TestCheckResourceAttr(ResourceType+"."+dateBasedRuleResourceLabel, "condition.0.relative_minutes_to_invocation", fmt.Sprintf("%d", relativeMinutesToInvocation2)),
				),
			},
			{
				ResourceName:      ResourceType + "." + dateBasedRuleResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyTaskManagementDateBasedRuleDestroyed,
	})
}

func testVerifyTaskManagementDateBasedRuleDestroyed(state *terraform.State) error {
	taskManagementApi := platformclientv2.NewTaskManagementApi()
	for _, res := range state.RootModule().Resources {
		if res.Type != ResourceType {
			continue
		}

		worktypeId, dateBasedRuleId := splitWorktypeBasedTerraformId(res.Primary.ID)
		dateBasedRule, resp, err := taskManagementApi.GetTaskmanagementWorktypeFlowsDatebasedRule(worktypeId, dateBasedRuleId)
		if dateBasedRule != nil {
			return fmt.Errorf("task management datebased rule (%s) still exists", res.Primary.ID)
		} else if util.IsStatus404(resp) {
			// Worktype no found, as expected
			continue
		} else {
			return fmt.Errorf("unexpected error: %s", err)
		}
	}

	// All datebased rules deleted
	return nil
}
