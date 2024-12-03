package task_management_workbin

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

/*
Test Class for the task management workbin Data Source
*/

func TestAccDataSourceTaskManagementWorkbin(t *testing.T) {
	t.Parallel()
	var (
		workbinResourceLabel = "workbin_1"
		workbinName          = "tf_workbin_" + uuid.NewString()
		workDescription      = "created for CX as Code test case"

		workbinDataSourceLabel = "workbin_data_source_1"
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateWorkbinResource(workbinResourceLabel, workbinName, workDescription, nullValue) +
					generateWorkbinDataSource(workbinDataSourceLabel, workbinName, ResourceType+"."+workbinResourceLabel),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data."+ResourceType+"."+workbinDataSourceLabel, "id", ResourceType+"."+workbinResourceLabel, "id"),
				),
			},
		},
	})
}

func generateWorkbinDataSource(dataSourceLabel string, name string, dependsOnResource string) string {
	return fmt.Sprintf(`data "%s" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, ResourceType, dataSourceLabel, name, dependsOnResource)
}
