package business_rules_schema

import (
	"fmt"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

/*
Test Class for the business rules schema Data Source
*/

func TestAccDataSourceBusinessRulesSchema(t *testing.T) {
	t.Parallel()

	if !businessRulesSchemaFtIsEnabled() {
		t.Skip("Skipping test as business rules schema feature toggle is not enabled")
		return
	}

	var (
		schemaResourceLabel = "schema_1"
		schemaName          = "tf_schema_" + uuid.NewString()
		schemaDescription   = "created for CX as Code test case"

		schemaDataSourceLabel = "business_rules_schema_data_source_1"
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateBusinessRulesSchemaResourceBasic(schemaResourceLabel, schemaName, schemaDescription) +
					generateBusinessRulesSchemaDataSource(schemaDataSourceLabel, schemaName, ResourceType+"."+schemaResourceLabel),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data."+ResourceType+"."+schemaDataSourceLabel, "id", ResourceType+"."+schemaResourceLabel, "id"),
				),
			},
		},
	})
}

func generateBusinessRulesSchemaDataSource(dataSourceLabel string, name string, dependsOnResource string) string {
	return fmt.Sprintf(`data "%s" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, ResourceType, dataSourceLabel, name, dependsOnResource)
}
