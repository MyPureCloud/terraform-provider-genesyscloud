package integration_credential

import (
	"fmt"
	"strconv"
	"testing"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
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
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateCredentialResource(
					credResource1,
					strconv.Quote(credName1),
					strconv.Quote(typeName1),
					GenerateCredentialFields(
						gcloud.GenerateMapProperty(key1, strconv.Quote(val1)),
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
						gcloud.GenerateMapProperty(key1, strconv.Quote(val1_2)),
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
						gcloud.GenerateMapProperty(key1, strconv.Quote(val1)),
						gcloud.GenerateMapProperty(key2, strconv.Quote(val2)),
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
						gcloud.GenerateMapProperty(key1, strconv.Quote(val1)),
						gcloud.GenerateMapProperty(key2, strconv.Quote(val2_2)),
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
						gcloud.GenerateMapProperty(key3, strconv.Quote(val3)),
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
						gcloud.GenerateMapProperty(key3, strconv.Quote(val3)),
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

func testVerifyCredentialDestroyed(state *terraform.State) error {
	integrationAPI := platformclientv2.NewIntegrationsApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_integration_credential" {
			continue
		}

		credential, resp, err := integrationAPI.GetIntegrationsCredential(rs.Primary.ID)
		if credential != nil {
			return fmt.Errorf("Credential (%s) still exists", rs.Primary.ID)
		} else if gcloud.IsStatus404(resp) {
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
