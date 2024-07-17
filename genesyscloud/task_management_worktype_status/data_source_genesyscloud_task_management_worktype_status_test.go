package task_management_worktype_status

import (
	"fmt"
	"github.com/google/uuid"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	workbin "terraform-provider-genesyscloud/genesyscloud/task_management_workbin"
	workitemSchema "terraform-provider-genesyscloud/genesyscloud/task_management_workitem_schema"
	workType "terraform-provider-genesyscloud/genesyscloud/task_management_worktype"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

/*
Test Class for the task management worktype status Data Source
*/

func TestAccDataSourceTaskManagementWorktypeStatus(t *testing.T) {
	t.Parallel()
	var (
		// Workbin
		wbResourceId  = "workbin_1"
		wbName        = "wb_" + uuid.NewString()
		wbDescription = "workbin created for CX as Code test case"

		// Schema
		wsResourceId  = "schema_1"
		wsName        = "ws_" + uuid.NewString()
		wsDescription = "workitem schema created for CX as Code test case"

		// Worktype
		wtResourceId  = "worktype_id"
		wtName        = "wt_" + uuid.NewString()
		wtDescription = "test worktype description"

		// Status Resource
		statusResourceId = "status_resource"
		statusName       = "status-" + uuid.NewString()
		statusCategory   = "Open"

		// Status Data Source
		statusDataSourceId = "status_data"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: workbin.GenerateWorkbinResource(wbResourceId, wbName, wbDescription, util.NullValue) +
					workitemSchema.GenerateWorkitemSchemaResourceBasic(wsResourceId, wsName, wsDescription) +
					workType.GenerateWorktypeResourceBasic(
						wtResourceId,
						wtName,
						wtDescription,
						fmt.Sprintf("genesyscloud_task_management_workbin.%s.id", wbResourceId),
						fmt.Sprintf("genesyscloud_task_management_workitem_schema.%s.id", wsResourceId),
						"",
					) +
					GenerateWorktypeStatusResource(
						statusResourceId,
						fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResourceId),
						statusName,
						statusCategory,
						"",
						util.NullValue,
						"",
					) +
					generateWorktypeStatusDataSource(
						statusDataSourceId,
						fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResourceId),
						statusName,
						resourceName+"."+statusResourceId,
					),
				Check: resource.ComposeTestCheckFunc(
					ValidateStatusIds(
						fmt.Sprintf("data.%s.%s", resourceName, statusDataSourceId), "id", fmt.Sprintf("%s.%s", resourceName, statusResourceId), "id",
					),
				),
			},
		},
	})
}

func generateWorktypeStatusDataSource(dataSourceId string, worktypeId string, name string, dependsOnResource string) string {
	return fmt.Sprintf(`data "%s" "%s" {
		worktype_id = %s
		name = "%s"
		depends_on=[%s]
	}
	`, resourceName, dataSourceId, worktypeId, name, dependsOnResource)
}
