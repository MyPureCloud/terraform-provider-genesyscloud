package task_management_worktype_flow_oncreate_rule

import (
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	workbin "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_workbin"
	workType "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_worktype"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

/*
Test Class for the task management oncreate rule Data Source
*/

func TestAccDataSourceTaskManagementOnCreateRule(t *testing.T) {
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

		// OnCreate Data Source
		onCreateRuleDataSourceLabel = "oncreate_rule_data"
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
					GenerateOnCreateRuleResource(
						onCreateRuleResourceLabel,
						fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResourceLabel),
						onCreateRuleName,
						"",
					) +
					generateOnCreateRuleDataSource(
						onCreateRuleDataSourceLabel,
						fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResourceLabel),
						onCreateRuleName,
						ResourceType+"."+onCreateRuleResourceLabel,
					),
				Check: resource.ComposeTestCheckFunc(
					validateRuleIds(
						"data."+ResourceType+"."+onCreateRuleDataSourceLabel, "id", ResourceType+"."+onCreateRuleResourceLabel, "id",
					),
				),
			},
			{
				ResourceName:      ResourceType + "." + onCreateRuleResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func generateOnCreateRuleDataSource(dataSourceLabel string, worktypeId string, name string, dependsOnResource string) string {
	return fmt.Sprintf(`data "%s" "%s" {
		worktype_id = %s
		name = "%s"
		depends_on=[%s]
	}
	`, ResourceType, dataSourceLabel, worktypeId, name, dependsOnResource)
}
