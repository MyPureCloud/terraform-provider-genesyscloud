package integration_action

import (
	"fmt"
	"strconv"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	integration "terraform-provider-genesyscloud/genesyscloud/integration"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

/*
Test Class for the Integration Actions Data Source
*/
func TestAccDataSourceIntegrationAction(t *testing.T) {
	var (
		integResource1  = "test_integration1"
		integTypeID     = "purecloud-data-actions"
		actionResource1 = "test-action1"
		actionResource2 = "test-action2"
		actionName1     = "Terraform Action1-" + uuid.NewString()
		actionCateg1    = "Genesys Cloud Data Actions"
		inputAttr1      = "service"
		outputAttr1     = "status"
		reqUrlTemplate1 = "/api/v2/users"
		reqType1        = "GET"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create without config
				Config: integration.GenerateIntegrationResource(
					integResource1,
					util.NullValue,
					strconv.Quote(integTypeID),
					// No config block
				) + generateIntegrationActionResource(
					actionResource1,
					actionName1,
					actionCateg1,
					"genesyscloud_integration."+integResource1+".id",
					util.NullValue, // Secure default (false)
					util.NullValue, // Timeout default
					util.GenerateJsonSchemaDocStr(inputAttr1),  // contract_input
					util.GenerateJsonSchemaDocStr(outputAttr1), // contract_output
					generateIntegrationActionConfigRequest(
						reqUrlTemplate1,
						reqType1,
						util.NullValue, // Default req templatezz
						"",             // No headers
					),
					// Default config response
				) + generateIntegrationActionDataSource(
					actionResource2,
					actionName1,
					"genesyscloud_integration_action."+actionResource1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_integration_action."+actionResource2, "id", "genesyscloud_integration_action."+actionResource1, "id"), // Default value would be "DISABLED"
				),
			},
		},
	})

}

func generateIntegrationActionDataSource(
	resourceID string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_integration_action" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceID, name, dependsOnResource)
}
