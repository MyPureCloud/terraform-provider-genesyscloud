package task_management_workitem_schema

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

/*
Test Class for the task management workitem schema Data Source
*/

func TestAccDataSourceTaskManagementWorkitemSchema(t *testing.T) {
	t.Parallel()
	var (
		schemaResId       = "schema_1"
		schemaName        = "tf_schema_" + uuid.NewString()
		schemaDescription = "created for CX as Code test case"

		schemaDataSourceId = "workitem_schema_data_source_1"
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateWorkitemSchemaResourceBasic(schemaResId, schemaName, schemaDescription) +
					generateWorkitemSchemaDataSource(schemaDataSourceId, schemaName, resourceName+"."+schemaResId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data."+resourceName+"."+schemaDataSourceId, "id", resourceName+"."+schemaResId, "id"),
				),
			},
		},
	})
}

func generateWorkitemSchemaDataSource(dataSourceId string, name string, dependsOnResource string) string {
	return fmt.Sprintf(`data "%s" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceName, dataSourceId, name, dependsOnResource)
}
