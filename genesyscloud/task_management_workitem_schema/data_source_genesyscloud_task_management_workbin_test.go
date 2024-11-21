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
		schemaResourceLabel = "schema_1"
		schemaName          = "tf_schema_" + uuid.NewString()
		schemaDescription   = "created for CX as Code test case"

		schemaDataSourceLabel = "workitem_schema_data_source_1"
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateWorkitemSchemaResourceBasic(schemaResourceLabel, schemaName, schemaDescription) +
					generateWorkitemSchemaDataSource(schemaDataSourceLabel, schemaName, ResourceType+"."+schemaResourceLabel),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data."+ResourceType+"."+schemaDataSourceLabel, "id", ResourceType+"."+schemaResourceLabel, "id"),
				),
			},
		},
	})
}

func generateWorkitemSchemaDataSource(dataSourceLabel string, name string, dependsOnResource string) string {
	return fmt.Sprintf(`data "%s" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, ResourceType, dataSourceLabel, name, dependsOnResource)
}
