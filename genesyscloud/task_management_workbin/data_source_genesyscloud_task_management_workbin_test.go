package task_management_workbin

import (
	"fmt"
	"testing"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

/*
Test Class for the task management workbin Data Source
*/

func TestAccDataSourceTaskManagementWorkbin(t *testing.T) {
	t.Parallel()
	var (
		workbinResId    = "workbin_1"
		workbinName     = "tf_workbin_" + uuid.NewString()
		workDescription = "created for CX as Code test case"

		workbinDataSourceId = "workbin_data_source_1"
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateWorkbinResource(workbinResId, workbinName, workDescription, nullValue) +
					generateWorkbinDataSource(workbinDataSourceId, workbinName, resourceName+"."+workbinResId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data."+resourceName+"."+workbinDataSourceId, "id", resourceName+"."+workbinResId, "id"),
				),
			},
		},
	})
}

func generateWorkbinDataSource(dataSourceId string, name string, dependsOnResource string) string {
	return fmt.Sprintf(`data "%s" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceName, dataSourceId, name, dependsOnResource)
}
