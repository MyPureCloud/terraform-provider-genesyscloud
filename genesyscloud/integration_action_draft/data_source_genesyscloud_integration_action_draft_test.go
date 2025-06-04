package integration_action_draft

import (
	"fmt"
	"strconv"
	integration "terraform-provider-genesyscloud/genesyscloud/integration"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

/*
Test Class for the Integration Actions Data Source
*/
func TestAccDataSourceIntegrationActionDraft(t *testing.T) {
	var (
		integResourceLabel1  = "test_integration1"
		integTypeID          = "purecloud-data-actions"
		actionResourceLabel1 = "test-action1"
		actionResourceLabel2 = "test-action2"
		actionName1          = "Terraform Action1-" + uuid.NewString()
		actionCateg1         = "Genesys Cloud Data Actions"
		inputAttr1           = "service"
		outputAttr1          = "status"
		reqUrlTemplate1      = "/api/v2/users"
		reqType1             = "GET"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create without config
				Config: integration.GenerateIntegrationResource(
					integResourceLabel1,
					util.NullValue,
					strconv.Quote(integTypeID),
					// No config block
				) + generateIntegrationActionDraftResource(
					actionResourceLabel1,
					actionName1,
					actionCateg1,
					"genesyscloud_integration."+integResourceLabel1+".id",
					util.NullValue, // Secure default (false)
					util.NullValue, // Timeout default
					util.GenerateJsonSchemaDocStr(inputAttr1),
					util.GenerateJsonSchemaDocStr(outputAttr1),
					generateIntegrationActionDraftConfigRequest(
						reqUrlTemplate1,
						reqType1,
						util.NullValue,
						"", // No headers
					),
					// Default config response
				) + generateIntegrationActionDataSource(
					actionResourceLabel2,
					actionName1,
					"genesyscloud_integration_action_draft."+actionResourceLabel1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_integration_action_draft."+actionResourceLabel2, "id", "genesyscloud_integration_action_draft."+actionResourceLabel1, "id"), // Default value would be "DISABLED"
				),
			},
		},
	})

}

func generateIntegrationActionDataSource(
	resourceLabel string,
	name string,
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_integration_action_draft" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceLabel, name, dependsOnResource)
}
