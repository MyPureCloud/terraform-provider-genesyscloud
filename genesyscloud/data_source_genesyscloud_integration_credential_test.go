package genesyscloud

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"strconv"
	"testing"
)

func TestAccDataSourceIntegrationCredential(t *testing.T) {

	var (
		credResource1 = "test_credential_1"
		credResource2 = "test_credential_2"
		credName1     = "Terraform Credential Test-" + uuid.NewString()
		//credName2     = "Terraform Credential Test-" + uuid.NewString()

		typeName1 = "basicAuth"
		//typeName2 = "callJourney"

		key1   = "userName"
		val1   = "someUserName"
		//val1_2 = "otherUserName"
		//key2   = "password"
		//val2   = "P@$$W0rd"
		//val2_2 = "$tr0ng3rP@$$W0rd"
		//
		//key3 = "authToken"
		//val3 = "fakeToken"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Create with config
				Config:  generateCredentialResource(
					credResource1,
					strconv.Quote(credName1),
					strconv.Quote(typeName1),
					generateCredentialFields(
						generateMapProperty(key1, strconv.Quote(val1)),
					),
				) + generateIntegrationCredentialDataSource(credResource2,
					credName1,
					"genesyscloud_integration_credential."+credResource1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_integration_credential."+credResource2, "id", "genesyscloud_integration_credential."+credResource1, "id"), // Default value would be "DISABLED"
				),
			},
		},
	})

}

func generateIntegrationCredentialDataSource(
	resourceID string,
	name string,
// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_integration_credential" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceID, name, dependsOnResource)
}
