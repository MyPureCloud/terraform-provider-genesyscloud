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
		wbResourceId  = "workbin_1"
		wbName        = "wb_" + uuid.NewString()
		wbDescription = "workbin created for CX as Code test case"

		wb2ResourceId  = "workbin_2"
		wb2Name        = "wb_" + uuid.NewString()
		wb2Description = "workbin created for CX as Code test case"

		// Schema
		wsResourceId  = "schema_1"
		wsName        = "ws_" + uuid.NewString()
		wsDescription = "workitem schema created for CX as Code test case"

		// worktype
		wtResName         = "tf_worktype_1"
		wtName            = "tf-worktype" + uuid.NewString()
		wtDescription     = "tf-worktype-description"

		// Worktype statuses
		statusResourceOpen   = "open-status"
		wtOStatusName        = "Open Status"
		wtOStatusDesc        = "Description of open status"
		wtOStatusCategory    = "Open"
		statusResourceClosed = "closed-status"
		wtCStatusName        = "Closed Status"
		wtCStatusDesc        = "Description of closed status"
		wtCStatusCategory    = "Closed"

		// basic workitem
		workitemRes = "workitem_1"
		workitem1   = workitemConfig{
			name:        "tf-workitem" + uuid.NewString(),
			worktype_id: fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResName),
		}

		workitemDataSrc = "workitem_1_data"

		taskMgmtConfig = workbin.GenerateWorkbinResource(wbResourceId, wbName, wbDescription, util.NullValue) +
			workbin.GenerateWorkbinResource(wb2ResourceId, wb2Name, wb2Description, util.NullValue) +
			workitemSchema.GenerateWorkitemSchemaResourceBasic(wsResourceId, wsName, wsDescription) +
			worktype.GenerateWorktypeResourceBasic(
				wtResName,
				wtName,
				wtDescription,
				fmt.Sprintf("genesyscloud_task_management_workbin.%s.id", wbResourceId),
				fmt.Sprintf("genesyscloud_task_management_workitem_schema.%s.id", wsResourceId),
				"",
			) +
			worktypeStatus.GenerateWorktypeStatusResource(
				statusResourceOpen,
				fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResName),
				wtOStatusName,
				wtOStatusCategory,
				wtOStatusDesc,
				util.NullValue,
				"",
			) +
			worktypeStatus.GenerateWorktypeStatusResource(
				statusResourceClosed,
				fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResName),
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
					generateWorkitemResourceBasic(workitemRes, workitem1.name, workitem1.worktype_id, "") +
					generateWorkitemDataSource(
						workitemDataSrc,
						workitem1.name,
						fmt.Sprintf("genesyscloud_task_management_workbin.%s.id", wbResourceId),
						"", // no worktype id filter
						fmt.Sprintf("genesyscloud_task_management_workitem.%s", workitemRes),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data."+resourceName+"."+workitemDataSrc, "id", resourceName+"."+workitemRes, "id"),
				),
			},
			// Test with using worktype id filter. API requires either or both workbin and worktype id filters.
			{
				Config: taskMgmtConfig +
					generateWorkitemResourceBasic(workitemRes, workitem1.name, workitem1.worktype_id, "") +
					generateWorkitemDataSource(
						workitemDataSrc,
						workitem1.name,
						"", // no workbin id filter
						fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResName),
						fmt.Sprintf("genesyscloud_task_management_workitem.%s", workitemRes),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data."+resourceName+"."+workitemDataSrc, "id", resourceName+"."+workitemRes, "id"),
				),
			},
		},
	})
}

func generateWorkitemDataSource(dataSourceId, name, workbinId, worktypeId, dependsOnResource string) string {
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
	`, resourceName, dataSourceId, name, additionalProps, dependsOnResource)
}
