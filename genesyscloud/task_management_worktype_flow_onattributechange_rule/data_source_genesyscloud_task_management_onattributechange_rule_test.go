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

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

/*
Test Class for the task management onattributechange rule Data Source
*/

func TestAccDataSourceTaskManagementOnAttributeChangeRule(t *testing.T) {
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
		statusResourceLabel = "status_1"
		statusName          = "status-" + uuid.NewString()
		statusCategory      = "InProgress"

		// OnAttributeChange Rule Resource
		onAttributeChangeRuleResourceLabel = "onattributechange_rule_resource"
		onAttributeChangeRuleName          = "onattributechange-" + uuid.NewString()
		onAttributeChangeRuleAttribute     = "statusId"

		// OnAttributeChange Data Source
		onAttributeChangeRuleDataSourceLabel = "onattributechange_rule_data"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: workbin.GenerateWorkbinResource(wbResourceLabel, wbName, wbDescription, util.NullValue) +
					workType.GenerateWorktypeResourceBasic(
						wtResourceLabel,
						wtName,
						wtDescription,
						fmt.Sprintf("genesyscloud_task_management_workbin.%s.id", wbResourceLabel),
						"",
					) +
					worktypeStatus.GenerateWorktypeStatusResource(
						statusResourceLabel,
						fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResourceLabel),
						statusName,
						statusCategory,
						"", util.NullValue, "",
					) +
					GenerateOnAttributeChangeRuleResource(
						onAttributeChangeRuleResourceLabel,
						fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResourceLabel),
						onAttributeChangeRuleName,
						onAttributeChangeRuleAttribute,
						fmt.Sprintf("genesyscloud_task_management_worktype_status.%s.id", statusResourceLabel),
						"", "",
					) +
					generateOnAttributeChangeRuleDataSource(
						onAttributeChangeRuleDataSourceLabel,
						fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResourceLabel),
						onAttributeChangeRuleName,
						ResourceType+"."+onAttributeChangeRuleResourceLabel,
					),
				Check: resource.ComposeTestCheckFunc(
					validateRuleIds(
						"data."+ResourceType+"."+onAttributeChangeRuleDataSourceLabel, "id", ResourceType+"."+onAttributeChangeRuleResourceLabel, "id",
					),
				),
			},
			{
				ResourceName:      ResourceType + "." + onAttributeChangeRuleResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func generateOnAttributeChangeRuleDataSource(dataSourceLabel string, worktypeId string, name string, dependsOnResource string) string {
	return fmt.Sprintf(`data "%s" "%s" {
		worktype_id = %s
		name = "%s"
		depends_on=[%s]
	}
	`, ResourceType, dataSourceLabel, worktypeId, name, dependsOnResource)
}
