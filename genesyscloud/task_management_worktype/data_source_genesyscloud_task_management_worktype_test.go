package task_management_worktype

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	workbin "terraform-provider-genesyscloud/genesyscloud/task_management_workbin"
	workitemSchema "terraform-provider-genesyscloud/genesyscloud/task_management_workitem_schema"
)

/*
Test Class for the task management worktype Data Source
*/

func TestAccDataSourceTaskManagementWorktype(t *testing.T) {
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
		wtRes = worktypeConfig{
			resID:            "worktype_1",
			name:             "tf_worktype_" + uuid.NewString(),
			description:      "worktype created for CX as Code test case",
			defaultWorkbinId: fmt.Sprintf("genesyscloud_task_management_workbin.%s.id", wbResourceId),
			schemaId:         fmt.Sprintf("genesyscloud_task_management_workitem_schema.%s.id", wsResourceId),
		}

		dataSourceId = "data_worktype_1"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: workbin.GenerateWorkbinResource(wbResourceId, wbName, wbDescription, util.NullValue) +
					workitemSchema.GenerateWorkitemSchemaResourceBasic(wsResourceId, wsName, wsDescription) +
					GenerateWorktypeResourceBasic(wtRes.resID, wtRes.name, wtRes.description, wtRes.defaultWorkbinId, wtRes.schemaId, "") +
					generateWorktypeDataSource(dataSourceId, wtRes.name, resourceName+"."+wtRes.resID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data."+resourceName+"."+dataSourceId, "id", resourceName+"."+wtRes.resID, "id"),
				),
			},
		},
	})
}

func generateWorktypeDataSource(dataSourceId string, name string, dependsOnResource string) string {
	return fmt.Sprintf(`data "%s" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceName, dataSourceId, name, dependsOnResource)
}
