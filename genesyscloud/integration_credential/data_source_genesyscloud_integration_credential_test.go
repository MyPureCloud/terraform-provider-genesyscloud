package integration_credential

import (
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

/*
Test Class for the Integration Credentials Data Source
*/
func TestAccDataSourceIntegrationCredential(t *testing.T) {
	var (
		credResourceLabel1 = "test_credential_1"
		credResourceLabel2 = "test_credential_2"
		credName1          = "Terraform Credential Test-" + uuid.NewString()
		typeName1          = "basicAuth"
		key1               = "userName"
		val1               = "someUserName"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					time.Sleep(30 * time.Second)
				},
				// Create
				Config: GenerateCredentialResource(
					credResourceLabel1,
					strconv.Quote(credName1),
					strconv.Quote(typeName1),
					GenerateCredentialFields(
						map[string]string{
							key1: strconv.Quote(val1),
						},
					),
				) + generateIntegrationCredentialDataSource(credResourceLabel2,
					credName1,
					"genesyscloud_integration_credential."+credResourceLabel1),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						time.Sleep(30 * time.Second) // Wait for 30 seconds for proper creation
						return nil
					},
					resource.TestCheckResourceAttrPair("data.genesyscloud_integration_credential."+credResourceLabel2, "id", "genesyscloud_integration_credential."+credResourceLabel1, "id"), // Default value would be "DISABLED"
				),
			},
		},
	})

}

func generateIntegrationCredentialDataSource(
	resourceLabel string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_integration_credential" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceLabel, name, dependsOnResource)
}
