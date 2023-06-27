package genesyscloud

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

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
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateCredentialResource(
					credResource1,
					strconv.Quote(credName1),
					strconv.Quote(typeName1),
					generateCredentialFields(
						generateMapProperty(key1, strconv.Quote(val1)),
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
				Config: generateCredentialResource(
					credResource1,
					strconv.Quote(credName2),
					strconv.Quote(typeName1),
					generateCredentialFields(
						generateMapProperty(key1, strconv.Quote(val1_2)),
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
				Config: generateCredentialResource(
					credResource1,
					strconv.Quote(credName2),
					strconv.Quote(typeName1),
					generateCredentialFields(
						generateMapProperty(key1, strconv.Quote(val1)),
						generateMapProperty(key2, strconv.Quote(val2)),
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
				Config: generateCredentialResource(
					credResource1,
					strconv.Quote(credName2),
					strconv.Quote(typeName1),
					generateCredentialFields(
						generateMapProperty(key1, strconv.Quote(val1)),
						generateMapProperty(key2, strconv.Quote(val2_2)),
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
				Config: generateCredentialResource(
					credResource2,
					strconv.Quote(credName1),
					strconv.Quote(typeName2),
					generateCredentialFields(
						generateMapProperty(key3, strconv.Quote(val3)),
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
				Config: generateCredentialResource(
					credResource2,
					strconv.Quote(credName2),
					strconv.Quote(typeName2),
					generateCredentialFields(
						generateMapProperty(key3, strconv.Quote(val3)),
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

func generateCredentialResource(resourceID string, name string, credentialType string, fields string) string {
	return fmt.Sprintf(`resource "genesyscloud_integration_credential" "%s" {
        name = %s
        credential_type_name = %s
        %s
	}
	`, resourceID, name, credentialType, fields)
}

func generateCredentialFields(fields ...string) string {
	return generateMapAttr("fields", fields...)
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
		} else if IsStatus404(resp) {
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
