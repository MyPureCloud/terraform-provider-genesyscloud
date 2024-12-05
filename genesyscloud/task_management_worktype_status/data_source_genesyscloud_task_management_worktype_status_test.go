package task_management_worktype_status

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	workbin "terraform-provider-genesyscloud/genesyscloud/task_management_workbin"
	workitemSchema "terraform-provider-genesyscloud/genesyscloud/task_management_workitem_schema"
	workType "terraform-provider-genesyscloud/genesyscloud/task_management_worktype"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

/*
Test Class for the task management worktype status Data Source
*/

func TestAccDataSourceTaskManagementWorktypeStatus(t *testing.T) {
	t.Parallel()
	var (
		// Workbin
		wbResourceLabel = "workbin_1"
		wbName          = "wb_" + uuid.NewString()
		wbDescription   = "workbin created for CX as Code test case"

		// Schema
		wsResourceLabel = "schema_1"
		wsName          = "ws_" + uuid.NewString()
		wsDescription   = "workitem schema created for CX as Code test case"

		// Worktype
		wtResourceLabel = "worktype_id"
		wtName          = "wt_" + uuid.NewString()
		wtDescription   = "test worktype description"

		// Status Resource
		statusResourceLabel = "status_resource"
		statusName          = "status-" + uuid.NewString()
		statusCategory      = "Open"

		// Status Data Source
		statusDataSourceLabel = "status_data"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: workbin.GenerateWorkbinResource(wbResourceLabel, wbName, wbDescription, util.NullValue) +
					workitemSchema.GenerateWorkitemSchemaResourceBasic(wsResourceLabel, wsName, wsDescription) +
					workType.GenerateWorktypeResourceBasic(
						wtResourceLabel,
						wtName,
						wtDescription,
						fmt.Sprintf("genesyscloud_task_management_workbin.%s.id", wbResourceLabel),
						fmt.Sprintf("genesyscloud_task_management_workitem_schema.%s.id", wsResourceLabel),
						"",
					) +
					GenerateWorktypeStatusResource(
						statusResourceLabel,
						fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResourceLabel),
						statusName,
						statusCategory,
						"",
						util.NullValue,
						"",
					) +
					generateWorktypeStatusDataSource(
						statusDataSourceLabel,
						fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResourceLabel),
						statusName,
						ResourceType+"."+statusResourceLabel,
					),
				Check: resource.ComposeTestCheckFunc(
					ValidateStatusIds(
						fmt.Sprintf("data.%s.%s", ResourceType, statusDataSourceLabel), "id", fmt.Sprintf("%s.%s", ResourceType, statusResourceLabel), "id",
					),
				),
			},
		},
	})
}

func generateWorktypeStatusDataSource(dataSourceLabel string, worktypeId string, name string, dependsOnResource string) string {
	return fmt.Sprintf(`data "%s" "%s" {
		worktype_id = %s
		name = "%s"
		depends_on=[%s]
	}
	`, ResourceType, dataSourceLabel, worktypeId, name, dependsOnResource)
}
