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
		wbResourceLabel = "workbin_1"
		wbName          = "wb_" + uuid.NewString()
		wbDescription   = "workbin created for CX as Code test case"

		// Schema
		wsResourceLabel = "schema_1"
		wsName          = "ws_" + uuid.NewString()
		wsDescription   = "workitem schema created for CX as Code test case"

		// Worktype
		wtRes = worktypeConfig{
			resourceLabel:    "worktype_1",
			name:             "tf_worktype_" + uuid.NewString(),
			description:      "worktype created for CX as Code test case",
			defaultWorkbinId: fmt.Sprintf("genesyscloud_task_management_workbin.%s.id", wbResourceLabel),
			schemaId:         fmt.Sprintf("genesyscloud_task_management_workitem_schema.%s.id", wsResourceLabel),
		}

		dataSourceLabel = "data_worktype_1"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: workbin.GenerateWorkbinResource(wbResourceLabel, wbName, wbDescription, util.NullValue) +
					workitemSchema.GenerateWorkitemSchemaResourceBasic(wsResourceLabel, wsName, wsDescription) +
					GenerateWorktypeResourceBasic(wtRes.resourceLabel, wtRes.name, wtRes.description, wtRes.defaultWorkbinId, wtRes.schemaId, "") +
					generateWorktypeDataSource(dataSourceLabel, wtRes.name, ResourceType+"."+wtRes.resourceLabel),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data."+ResourceType+"."+dataSourceLabel, "id", ResourceType+"."+wtRes.resourceLabel, "id"),
				),
			},
		},
	})
}

func generateWorktypeDataSource(dataSourceLabel string, name string, dependsOnResource string) string {
	return fmt.Sprintf(`data "%s" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, ResourceType, dataSourceLabel, name, dependsOnResource)
}
