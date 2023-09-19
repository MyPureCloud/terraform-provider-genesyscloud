package integration_credential

import (
	"fmt"
	"strconv"
	"testing"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

/*
Test Class for the Integration Credentials Data Source
*/
func TestAccDataSourceIntegrationCredential(t *testing.T) {
	var (
		credResource1 = "test_credential_1"
		credResource2 = "test_credential_2"
		credName1     = "Terraform Credential Test-" + uuid.NewString()
		typeName1     = "basicAuth"
		key1          = "userName"
		val1          = "someUserName"
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
