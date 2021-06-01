package genesyscloud

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/platformclientv2"
)

func TestAccResourceCredential(t *testing.T) {
	var (
		credResource1 = "test_credential_1"
		credResource2 = "test_credential_2"
		credName1     = "Terraform Credential Test-" + uuid.NewString()
		credName2     = "Terraform Credential Test-" + uuid.NewString()

		typeName1 = "basicAuth"
		typeName2 = "callJourney"

		key1 = "userName"
		val1 = "someUserName"

		key2 = "authToken"
		val2 = "fakeToken"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateCredentialResource(
					credResource1,
					strconv.Quote(credName1),
					strconv.Quote(typeName1),
					generateCredentialFields(
						generateMapKeyValue(key1, strconv.Quote(val1)),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_credential."+credResource1, "name", credName1),
					resource.TestCheckResourceAttr("genesyscloud_credential."+credResource1, "credential_type_name", typeName1),
					resource.TestCheckNoResourceAttr("genesyscloud_credential."+credResource1, "fields"),
				),
			},
			{
				// Update name
				Config: generateCredentialResource(
					credResource1,
					strconv.Quote(credName2),
					strconv.Quote(typeName1),
					generateCredentialFields(
						generateMapKeyValue(key1, strconv.Quote(val1)),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_credential."+credResource1, "name", credName2),
					resource.TestCheckResourceAttr("genesyscloud_credential."+credResource1, "credential_type_name", typeName1),
					resource.TestCheckNoResourceAttr("genesyscloud_credential."+credResource1, "fields"),
				),
			},
			{
				// Import/Read
				ResourceName:            "genesyscloud_credential." + credResource1,
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
						generateMapKeyValue(key2, strconv.Quote(val2)),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_credential."+credResource2, "name", credName1),
					resource.TestCheckResourceAttr("genesyscloud_credential."+credResource2, "credential_type_name", typeName2),
					resource.TestCheckNoResourceAttr("genesyscloud_credential."+credResource2, "fields"),
				),
			},
			{
				// Update name
				Config: generateCredentialResource(
					credResource2,
					strconv.Quote(credName2),
					strconv.Quote(typeName2),
					generateCredentialFields(
						generateMapKeyValue(key2, strconv.Quote(val2)),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_credential."+credResource2, "name", credName2),
					resource.TestCheckResourceAttr("genesyscloud_credential."+credResource2, "credential_type_name", typeName2),
					resource.TestCheckNoResourceAttr("genesyscloud_credential."+credResource2, "fields"),
				),
			},
			{
				// Import/Read
				ResourceName:            "genesyscloud_credential." + credResource2,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"fields"},
			},
		},
		CheckDestroy: testVerifyCredentialDestroyed,
	})
}

func generateCredentialResource(resourceID string, name string, credentialType string, fields string) string {
	return fmt.Sprintf(`resource "genesyscloud_credential" "%s" {
        name = %s
        credential_type_name = %s
        %s
	}
	`, resourceID, name, credentialType, fields)
}

func generateCredentialFields(fields ...string) string {
	return fmt.Sprintf(`fields = {
        %s
	}
	`, strings.Join(fields, "\n"))
}

func generateMapKeyValue(key string, val string) string {
	return fmt.Sprintf(`
        %s = %s
    `, key, val)
}

func testVerifyCredentialDestroyed(state *terraform.State) error {
	integrationAPI := platformclientv2.NewIntegrationsApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_credential" {
			continue
		}

		credential, resp, err := integrationAPI.GetIntegrationsCredential(rs.Primary.ID)
		if credential != nil {
			return fmt.Errorf("Credential (%s) still exists", rs.Primary.ID)
		} else if resp != nil && resp.StatusCode == 404 {
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
