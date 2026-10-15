package integration

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

/*
Test Class for the Integrations Data Source
*/
func TestAccDataSourceIntegration(t *testing.T) {

	var (
		inteResourceLabel1 = "test_integration1"
		inteResourceLabel2 = "test_integration2"
		inteName1          = "Terraform Integration Test-" + uuid.NewString()
		typeID             = "embedded-client-app"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create with config
				Config: GenerateIntegrationResource(
					inteResourceLabel1,
					util.NullValue, //Empty intended_state, default value is "DISABLED"
					strconv.Quote(typeID),
					GenerateIntegrationConfig(
						strconv.Quote(inteName1),
						util.NullValue, //Empty notes
						"",             //Empty credential ID
						util.NullValue, //Empty properties
						util.NullValue, //Empty advanced JSON
					),
					// No config block
				) + generateIntegrationDataSource(inteResourceLabel2,
					inteName1,
					"genesyscloud_integration."+inteResourceLabel1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_integration."+inteResourceLabel2, "id", "genesyscloud_integration."+inteResourceLabel1, "id"), // Default value would be "DISABLED"
				),
			},
			{
				// Look up by name + integration_type
				Config: GenerateIntegrationResource(
					inteResourceLabel1,
					util.NullValue, //Empty intended_state, default value is "DISABLED"
					strconv.Quote(typeID),
					GenerateIntegrationConfig(
						strconv.Quote(inteName1),
						util.NullValue, //Empty notes
						"",             //Empty credential ID
						util.NullValue, //Empty properties
						util.NullValue, //Empty advanced JSON
					),
				) + generateIntegrationDataSourceWithType(inteResourceLabel2,
					inteName1,
					typeID,
					"genesyscloud_integration."+inteResourceLabel1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_integration."+inteResourceLabel2, "id", "genesyscloud_integration."+inteResourceLabel1, "id"),
					resource.TestCheckResourceAttr("data.genesyscloud_integration."+inteResourceLabel2, "integration_type", typeID),
				),
			},
		},
	})

}

func generateIntegrationDataSource(
	resourceLabel string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_integration" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceLabel, name, dependsOnResource)
}

func generateIntegrationDataSourceWithType(
	resourceLabel string,
	name string,
	integrationType string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_integration" "%s" {
		name = "%s"
		integration_type = "%s"
		depends_on=[%s]
	}
	`, resourceLabel, name, integrationType, dependsOnResource)
}
