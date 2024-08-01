package integration_custom_auth_action

import (
	"fmt"
	"strconv"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
	"time"

	integration "terraform-provider-genesyscloud/genesyscloud/integration"
	integrationCred "terraform-provider-genesyscloud/genesyscloud/integration_credential"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

/*
Test Class for the Integration Custom Auth Actions Data Source
*/
func TestAccDataSourceIntegrationCustomAuthAction(t *testing.T) {
	var (
		// Integration Credentials
		credentialResource1      = "test_integration_credential_1"
		credentialResourceName   = "Terraform Cred-" + uuid.NewString()
		credKey1                 = "loginUrl"
		credVal1                 = "https://www.test-login.com"
		credentialResourceConfig = integrationCred.GenerateCredentialResource(
			credentialResource1,
			strconv.Quote(credentialResourceName),
			strconv.Quote(customAuthCredentialType),
			integrationCred.GenerateCredentialFields(
				map[string]string{credKey1: strconv.Quote(credVal1)},
			),
		)

		// Web Services Data Action Integration
		integResource1            = "test_integration1"
		integResourceName1        = "Terraform Integration-" + uuid.NewString()
		integTypeID               = "custom-rest-actions"
		integrationResourceConfig = integration.GenerateIntegrationResource(
			integResource1,
			util.NullValue,
			strconv.Quote(integTypeID),
			integration.GenerateIntegrationConfig(
				strconv.Quote(integResourceName1),
				util.NullValue, // no notes
				fmt.Sprintf("basicAuth = genesyscloud_integration_credential.%s.id", credentialResource1),
				util.NullValue, // no properties
				util.NullValue, // no advanced properties
			),
		)

		// Data Source
		customAuthSource = "custom-auth-1"
		dataSourceConfig = generateCustomAuthActionDataSource(customAuthSource, "genesyscloud_integration."+integResource1+".id", "genesyscloud_integration."+integResource1)

		config = credentialResourceConfig + integrationResourceConfig + dataSourceConfig
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckCustomAuthId("data.genesyscloud_integration_custom_auth_action."+customAuthSource, "genesyscloud_integration."+integResource1),
					func(s *terraform.State) error {
						time.Sleep(30 * time.Second) // Wait for 30 seconds for proper deletion
						return nil
					},
				),
			},
		},
	})

}

func generateCustomAuthActionDataSource(resourceID string, integrationId string, dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_integration_custom_auth_action" "%s" {
		parent_integration_id = %s
		depends_on=[%s]
	}
	`, resourceID, integrationId, dependsOnResource)
}

// testCheckCustomAuthId verified if the ID of the data source matches the expected custom auth id
// from the specified integration resource
func testCheckCustomAuthId(authSourceResName string, integrationResName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		integrationResource, ok := state.RootModule().Resources[integrationResName]
		if !ok {
			return fmt.Errorf("failed to find integration %s in state", integrationResName)
		}
		authDataSource, ok := state.RootModule().Resources[authSourceResName]
		if !ok {
			return fmt.Errorf("failed to find auth data source %s in state", integrationResName)
		}

		expectedAuthId := getCustomAuthIdFromIntegration(integrationResource.Primary.ID)
		if authDataSource.Primary.ID != expectedAuthId {
			return fmt.Errorf("integration %s expected auth id %s does not match actual: %s", integrationResName, expectedAuthId, authDataSource.Primary.ID)
		}

		return nil
	}
}
