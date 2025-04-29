package task_management_worktype_flow_datebased_rule

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
Test Class for the task management datebased rule Data Source
*/

func TestAccDataSourceTaskManagementDateBasedRule(t *testing.T) {
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
		dateBasedRuleResourceLabel  = "datebased_rule_resource"
		dateBasedRuleName           = "datebased-" + uuid.NewString()
		dateBasedRuleAttribute      = "dateDue"
		relativeMinutesToInvocation = 30

		// DateBased Data Source
		dateBasedRuleDataSourceLabel = "datebased_rule_data"
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
					GenerateDateBasedRuleResource(
						dateBasedRuleResourceLabel,
						fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResourceLabel),
						dateBasedRuleName,
						dateBasedRuleAttribute,
						relativeMinutesToInvocation,
					) +
					generateDateBasedRuleDataSource(
						dateBasedRuleDataSourceLabel,
						fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResourceLabel),
						dateBasedRuleName,
						ResourceType+"."+dateBasedRuleResourceLabel,
					),
				Check: resource.ComposeTestCheckFunc(
					validateRuleIds(
						"data."+ResourceType+"."+dateBasedRuleDataSourceLabel, "id", ResourceType+"."+dateBasedRuleResourceLabel, "id",
					),
				),
			},
			{
				ResourceName:      ResourceType + "." + dateBasedRuleResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func generateDateBasedRuleDataSource(dataSourceLabel string, worktypeId string, name string, dependsOnResource string) string {
	return fmt.Sprintf(`data "%s" "%s" {
		worktype_id = %s
		name = "%s"
		depends_on=[%s]
	}
	`, ResourceType, dataSourceLabel, worktypeId, name, dependsOnResource)
}
