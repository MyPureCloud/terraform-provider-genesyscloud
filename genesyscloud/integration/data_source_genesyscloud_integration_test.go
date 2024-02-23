package integration

import (
	"fmt"
	"strconv"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

/*
Test Class for the Integrations Data Source
*/
func TestAccDataSourceIntegration(t *testing.T) {

	var (
		inteResource1 = "test_integration1"
		inteResource2 = "test_integration2"
		inteName1     = "Terraform Integration Test-" + uuid.NewString()
		typeID        = "embedded-client-app"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create with config
				Config: GenerateIntegrationResource(
					inteResource1,
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
				) + generateIntegrationDataSource(inteResource2,
					inteName1,
					"genesyscloud_integration."+inteResource1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_integration."+inteResource2, "id", "genesyscloud_integration."+inteResource1, "id"), // Default value would be "DISABLED"
				),
			},
		},
	})

}

func generateIntegrationDataSource(
	resourceID string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_integration" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceID, name, dependsOnResource)
}
