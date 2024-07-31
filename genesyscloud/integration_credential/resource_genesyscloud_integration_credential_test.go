package integration_credential

import (
	"fmt"
	"strconv"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The resource_genesyscloud_integration_credential_test.go contains all of the test cases for running the resource
tests for integration_credentials.
*/
func TestAccResourceCredential(t *testing.T) {
	var (
		credResource1 = "test_credential_1"
		credResource2 = "test_credential_2"
		credName1     = "Terraform Credential Test-" + uuid.NewString()
		credName2     = "Terraform Credential Test-" + uuid.NewString()

		typeName1 = "basicAuth"
		typeName2 = "callJourney"

		key1   = "userName"
		val1   = "someUserName"
		val1_2 = "otherUserName"
		key2   = "password"
		val2   = "P@$$W0rd"
		val2_2 = "$tr0ng3rP@$$W0rd"

		key3 = "authToken"
		val3 = "fakeToken"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateCredentialResource(
					credResource1,
					strconv.Quote(credName1),
					strconv.Quote(typeName1),
					GenerateCredentialFields(
						map[string]string{
							key1: strconv.Quote(val1),
						},
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_integration_credential."+credResource1, "name", credName1),
					resource.TestCheckResourceAttr("genesyscloud_integration_credential."+credResource1, "credential_type_name", typeName1),
					resource.TestCheckResourceAttr("genesyscloud_integration_credential."+credResource1, "fields."+key1, val1),
				),
			},
			{
				// Update name and value of one field
				Config: GenerateCredentialResource(
					credResource1,
					strconv.Quote(credName2),
					strconv.Quote(typeName1),
					GenerateCredentialFields(
						map[string]string{
							key1: strconv.Quote(val1_2),
						},
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_integration_credential."+credResource1, "name", credName2),
					resource.TestCheckResourceAttr("genesyscloud_integration_credential."+credResource1, "credential_type_name", typeName1),
					resource.TestCheckResourceAttr("genesyscloud_integration_credential."+credResource1, "fields."+key1, val1_2),
				),
			},
			{
				// Add another field
				Config: GenerateCredentialResource(
					credResource1,
					strconv.Quote(credName2),
					strconv.Quote(typeName1),
					GenerateCredentialFields(
						map[string]string{
							key1: strconv.Quote(val1),
							key2: strconv.Quote(val2),
						},
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_integration_credential."+credResource1, "name", credName2),
					resource.TestCheckResourceAttr("genesyscloud_integration_credential."+credResource1, "credential_type_name", typeName1),
					resource.TestCheckResourceAttr("genesyscloud_integration_credential."+credResource1, "fields."+key1, val1),
					resource.TestCheckResourceAttr("genesyscloud_integration_credential."+credResource1, "fields."+key2, val2),
				),
			},
			{
				// Update second field
				Config: GenerateCredentialResource(
					credResource1,
					strconv.Quote(credName2),
					strconv.Quote(typeName1),
					GenerateCredentialFields(
						map[string]string{
							key1: strconv.Quote(val1),
							key2: strconv.Quote(val2_2),
						},
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_integration_credential."+credResource1, "name", credName2),
					resource.TestCheckResourceAttr("genesyscloud_integration_credential."+credResource1, "credential_type_name", typeName1),
					resource.TestCheckResourceAttr("genesyscloud_integration_credential."+credResource1, "fields."+key1, val1),
					resource.TestCheckResourceAttr("genesyscloud_integration_credential."+credResource1, "fields."+key2, val2_2),
				),
			},
			{
				// Import/Read
				ResourceName:            "genesyscloud_integration_credential." + credResource1,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"fields"},
			},
			{
				// Create another resource with different type
				Config: GenerateCredentialResource(
					credResource2,
					strconv.Quote(credName1),
					strconv.Quote(typeName2),
					GenerateCredentialFields(
						map[string]string{
							key3: strconv.Quote(val3),
						},
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_integration_credential."+credResource2, "name", credName1),
					resource.TestCheckResourceAttr("genesyscloud_integration_credential."+credResource2, "credential_type_name", typeName2),
					resource.TestCheckResourceAttr("genesyscloud_integration_credential."+credResource2, "fields."+key3, val3),
				),
			},
			{
				// Update name
				Config: GenerateCredentialResource(
					credResource2,
					strconv.Quote(credName2),
					strconv.Quote(typeName2),
					GenerateCredentialFields(
						map[string]string{
							key3: strconv.Quote(val3),
						},
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_integration_credential."+credResource2, "name", credName2),
					resource.TestCheckResourceAttr("genesyscloud_integration_credential."+credResource2, "credential_type_name", typeName2),
					resource.TestCheckResourceAttr("genesyscloud_integration_credential."+credResource2, "fields."+key3, val3),
				),
			},
			{
				// Import/Read
				ResourceName:            "genesyscloud_integration_credential." + credResource2,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"fields"},
			},
		},
		CheckDestroy: testVerifyCredentialDestroyed,
	})
}

// Resource Credential DEVTOOLING-448
// This tests to make sure that we can successfully create an integration credential for a Genesys Cloud oauth client without providing a client secret
func TestAccGenesysCloudOAuthResourceCredential(t *testing.T) {
	var (
		oAuthResourceID = "test_genesys_oauth_client"
		oAuthName       = "test_genesys_oauth_client" + uuid.NewString()

		credResourceID = "test_genesys_oauth_integration_cred"
		credName       = "Terraform Genesys Oauth Credential Test-" + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateGenesysOauthCredentialResource(oAuthResourceID, oAuthName) + " " + generateOAuthIntegrationCredentialResource(credResourceID, credName, oAuthResourceID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_integration_credential."+credResourceID, "name", credName),
					resource.TestCheckResourceAttr("genesyscloud_oauth_client."+oAuthResourceID, "name", oAuthName),
				),
			},
		},
		CheckDestroy: testVerifyCredentialDestroyed,
	})
}

// TestAccGenesysCloudOAuthResourceCredentialWithSecret will check to make sure we can still create Genesys Cloud
// integration credential by providing the oauth client and id and secret.  This is how we would normally do it.
func TestAccGenesysCloudOAuthResourceCredentialWithSecret(t *testing.T) {
	var (
		credResource = "test_genesyscloud_oauth_integration_credential_1"
		credName     = "Terraform Oauth Integration Credential Test-" + uuid.NewString()

		typeName = "pureCloudOAuthClient"

		clientId     = uuid.NewString()
		clientSecret = uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateCredentialResource(
					credResource,
					strconv.Quote(credName),
					strconv.Quote(typeName),
					GenerateCredentialFields(
						map[string]string{
							"clientId":     strconv.Quote(clientId),
							"clientSecret": strconv.Quote(clientSecret),
						},
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_integration_credential."+credResource, "name", credName),
					resource.TestCheckResourceAttr("genesyscloud_integration_credential."+credResource, "credential_type_name", typeName),
					resource.TestCheckResourceAttr("genesyscloud_integration_credential."+credResource, "fields.clientId", clientId),
					resource.TestCheckResourceAttr("genesyscloud_integration_credential."+credResource, "fields.clientSecret", clientSecret),
				),
			},
		},
		CheckDestroy: testVerifyCredentialDestroyed,
	})
}

func testVerifyCredentialDestroyed(state *terraform.State) error {
	integrationAPI := platformclientv2.NewIntegrationsApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_integration_credential" {
			continue
		}

		credential, resp, err := integrationAPI.GetIntegrationsCredential(rs.Primary.ID)
		if credential != nil {
			return fmt.Errorf("Credential (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// Credential not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All credentials destroyed
	return nil
}

// These two methods are used to test generate a Genesys Cloud OAuth Client so we can test thing the OAuth Client Caching
//introduce as part of DevTooling-448

func generateOAuthIntegrationCredentialResource(resourceID string, name string, oauthClientResourceID string) string {
	return fmt.Sprintf(`resource "genesyscloud_integration_credential" "%s" {
			name                 = "%s"
            credential_type_name = "pureCloudOAuthClient"
            fields = {
    			clientId = "${genesyscloud_oauth_client.%s.id}"
            }
    }`, resourceID, name, oauthClientResourceID)
}

func generateGenesysOauthCredentialResource(resourceID string, name string) string {

	return fmt.Sprintf(`
      data "genesyscloud_auth_role" "admin" {
		name = "Admin"
	  }

      resource "genesyscloud_oauth_client" "%s" {
		name                          =  "%s"
		description                   = "A Genesys Cloud OAuth Client used to test caching logic from 448"
		authorized_grant_type         = "CLIENT-CREDENTIALS"
		state                         = "active"


		roles {
			role_id     = data.genesyscloud_auth_role.admin.id
		}
      }
	`, resourceID, name)
}
