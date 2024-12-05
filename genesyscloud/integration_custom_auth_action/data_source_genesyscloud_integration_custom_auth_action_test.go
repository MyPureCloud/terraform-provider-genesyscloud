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
		credentialResourceLabel1   = "test_integration_credential_1"
		credentialResourceTypeAttr = "Terraform Cred-" + uuid.NewString()
		credKey1                   = "loginUrl"
		credVal1                   = "https://www.test-login.com"
		credentialResourceConfig   = integrationCred.GenerateCredentialResource(
			credentialResourceLabel1,
			strconv.Quote(credentialResourceTypeAttr),
			strconv.Quote(customAuthCredentialType),
			integrationCred.GenerateCredentialFields(
				map[string]string{credKey1: strconv.Quote(credVal1)},
			),
		)

		// Web Services Data Action Integration
		integResourceLabel1       = "test_integration1"
		integResourceTypeAttr1    = "Terraform Integration-" + uuid.NewString()
		integTypeID               = "custom-rest-actions"
		integrationResourceConfig = integration.GenerateIntegrationResource(
			integResourceLabel1,
			util.NullValue,
			strconv.Quote(integTypeID),
			integration.GenerateIntegrationConfig(
				strconv.Quote(integResourceTypeAttr1),
				util.NullValue, // no notes
				fmt.Sprintf("basicAuth = genesyscloud_integration_credential.%s.id", credentialResourceLabel1),
				util.NullValue, // no properties
				util.NullValue, // no advanced properties
			),
		)

		// Data Source
		customAuthSource = "custom-auth-1"
		dataSourceConfig = generateCustomAuthActionDataSource(customAuthSource, "genesyscloud_integration."+integResourceLabel1+".id", "genesyscloud_integration."+integResourceLabel1)

		config = credentialResourceConfig + integrationResourceConfig + dataSourceConfig
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckCustomAuthId("data.genesyscloud_integration_custom_auth_action."+customAuthSource, "genesyscloud_integration."+integResourceLabel1),
					func(s *terraform.State) error {
						time.Sleep(30 * time.Second) // Wait for 30 seconds for proper deletion
						return nil
					},
				),
			},
		},
	})

}

func generateCustomAuthActionDataSource(dataSourceLabel string, integrationId string, dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_integration_custom_auth_action" "%s" {
		parent_integration_id = %s
		depends_on=[%s]
	}
	`, dataSourceLabel, integrationId, dependsOnResource)
}

// testCheckCustomAuthId verified if the ID of the data source matches the expected custom auth id
// from the specified integration resource
func testCheckCustomAuthId(authSourceResourcePath string, integrationResourcePath string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		integrationResource, ok := state.RootModule().Resources[integrationResourcePath]
		if !ok {
			return fmt.Errorf("failed to find integration %s in state", integrationResourcePath)
		}
		authDataSource, ok := state.RootModule().Resources[authSourceResourcePath]
		if !ok {
			return fmt.Errorf("failed to find auth data source %s in state", integrationResourcePath)
		}

		expectedAuthId := getCustomAuthIdFromIntegration(integrationResource.Primary.ID)
		if authDataSource.Primary.ID != expectedAuthId {
			return fmt.Errorf("integration %s expected auth id %s does not match actual: %s", integrationResourcePath, expectedAuthId, authDataSource.Primary.ID)
		}

		return nil
	}
}
