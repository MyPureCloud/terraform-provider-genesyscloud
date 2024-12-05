package task_management_workitem

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	workbin "terraform-provider-genesyscloud/genesyscloud/task_management_workbin"
	workitemSchema "terraform-provider-genesyscloud/genesyscloud/task_management_workitem_schema"
	worktype "terraform-provider-genesyscloud/genesyscloud/task_management_worktype"
	worktypeStatus "terraform-provider-genesyscloud/genesyscloud/task_management_worktype_status"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

/*
Test Class for the task management workitem Data Source
*/

func TestAccDataSourceTaskManagementWorkitem(t *testing.T) {
	t.Parallel()
	var (
		// Workbin
		wbResourceLabel = "workbin_1"
		wbName          = "wb_" + uuid.NewString()
		wbDescription   = "workbin created for CX as Code test case"

		wb2ResourceLabel = "workbin_2"
		wb2Name          = "wb_" + uuid.NewString()
		wb2Description   = "workbin created for CX as Code test case"

		// Schema
		wsResourceLabel = "schema_1"
		wsName          = "ws_" + uuid.NewString()
		wsDescription   = "workitem schema created for CX as Code test case"

		// worktype
		wtResourceLabel = "tf_worktype_1"
		wtName          = "tf-worktype" + uuid.NewString()
		wtDescription   = "tf-worktype-description"

		// Worktype statuses
		statusResourceLabelOpen   = "open-status"
		wtOStatusName             = "Open Status"
		wtOStatusDesc             = "Description of open status"
		wtOStatusCategory         = "Open"
		statusResourceLabelClosed = "closed-status"
		wtCStatusName             = "Closed Status"
		wtCStatusDesc             = "Description of closed status"
		wtCStatusCategory         = "Closed"

		// basic workitem
		workitemResourceLabel = "workitem_1"
		workitem1             = workitemConfig{
			name:        "tf-workitem" + uuid.NewString(),
			worktype_id: fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResourceLabel),
		}

		workitemDataSrc = "workitem_1_data"

		taskMgmtConfig = workbin.GenerateWorkbinResource(wbResourceLabel, wbName, wbDescription, util.NullValue) +
			workbin.GenerateWorkbinResource(wb2ResourceLabel, wb2Name, wb2Description, util.NullValue) +
			workitemSchema.GenerateWorkitemSchemaResourceBasic(wsResourceLabel, wsName, wsDescription) +
			worktype.GenerateWorktypeResourceBasic(
				wtResourceLabel,
				wtName,
				wtDescription,
				fmt.Sprintf("genesyscloud_task_management_workbin.%s.id", wbResourceLabel),
				fmt.Sprintf("genesyscloud_task_management_workitem_schema.%s.id", wsResourceLabel),
				"",
			) +
			worktypeStatus.GenerateWorktypeStatusResource(
				statusResourceLabelOpen,
				fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResourceLabel),
				wtOStatusName,
				wtOStatusCategory,
				wtOStatusDesc,
				util.NullValue,
				"",
			) +
			worktypeStatus.GenerateWorktypeStatusResource(
				statusResourceLabelClosed,
				fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResourceLabel),
				wtCStatusName,
				wtCStatusCategory,
				wtCStatusDesc,
				util.NullValue,
				"",
			)
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			// Test with using workbin id filter. API requires either or both workbin and worktype id filters.
			{
				Config: taskMgmtConfig +
					generateWorkitemResourceBasic(workitemResourceLabel, workitem1.name, workitem1.worktype_id, "") +
					generateWorkitemDataSource(
						workitemDataSrc,
						workitem1.name,
						fmt.Sprintf("genesyscloud_task_management_workbin.%s.id", wbResourceLabel),
						"", // no worktype id filter
						fmt.Sprintf("genesyscloud_task_management_workitem.%s", workitemResourceLabel),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data."+ResourceType+"."+workitemDataSrc, "id", ResourceType+"."+workitemResourceLabel, "id"),
				),
			},
			// Test with using worktype id filter. API requires either or both workbin and worktype id filters.
			{
				Config: taskMgmtConfig +
					generateWorkitemResourceBasic(workitemResourceLabel, workitem1.name, workitem1.worktype_id, "") +
					generateWorkitemDataSource(
						workitemDataSrc,
						workitem1.name,
						"", // no workbin id filter
						fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResourceLabel),
						fmt.Sprintf("genesyscloud_task_management_workitem.%s", workitemResourceLabel),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data."+ResourceType+"."+workitemDataSrc, "id", ResourceType+"."+workitemResourceLabel, "id"),
				),
			},
		},
	})
}

func generateWorkitemDataSource(dataSourceLabel, name, workbinId, worktypeId, dependsOnResource string) string {
	additionalProps := ""
	if workbinId != "" {
		additionalProps += fmt.Sprintf("workbin_id = %s\n", workbinId)
	}
	if worktypeId != "" {
		additionalProps += fmt.Sprintf("worktype_id = %s\n", worktypeId)
	}

	return fmt.Sprintf(`
	data "%s" "%s" {
		name = "%s"
		%s
		depends_on=[%s]
	}
	`, ResourceType, dataSourceLabel, name, additionalProps, dependsOnResource)
}
