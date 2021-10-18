package genesyscloud

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"strconv"
	"testing"
)

func TestAccDataSourceIntegration(t *testing.T) {

	var (
		inteResource1 = "test_integration1"
		inteResource2 = "test_integration2"
		inteName1     = "Terraform Integration Test-" + uuid.NewString()
		//inteName2     = "Terraform Integration Test-" + uuid.NewString()
		//
		//defaultState = "DISABLED"
		//enabledState = "ENABLED"
		//configNotes  = "some notes"
		//configNotes2 = "This is a note"

		typeID  = "embedded-client-app"
		//typeID2 = "custom-smtp-server"
		//
		//displayTypeKey  = "displayType"
		//sandboxKey      = "sandbox"
		//urlKey          = "url"
		//groupsKey       = "groups"
		//propDisplayType = "standalone"
		//propSandbox     = "allow-scripts,allow-same-origin,allow-forms,allow-modals"
		//propURL         = "https://mypurecloud.github.io/purecloud-premium-app/wizard/index.html"
		//groupResource1  = "test_group"
		//groupName       = "terraform integration test group-" + uuid.NewString()
		//fakeGroupID     = "123456789"
		//emptyJSON       = "{}"
		//
		//credResource1 = "test_credential"
		//credName1     = "Terraform Credential Test-" + uuid.NewString()
		//credTypeName1 = "basicAuth"
		//key1          = "userName"
		//val1          = "someUserName"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Create without config
				Config: generateIntegrationResource(
					inteResource1,
					nullValue, //Empty intended_state, default value is "DISABLED"
					strconv.Quote(typeID),
					// No config block
				) + generateIntegrationDataSource(inteResource2,
					inteName1,
					"genesyscloud_integration." + inteResource1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("genesyscloud_integration."+inteResource2, "id", "genesyscloud_integration."+inteResource2, "id"), // Default value would be "DISABLED"
				),
			},
		},
	})
}

func generateIntegrationDataSource(
	resourceID string,
	name string,
// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_integration" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceID, name, dependsOnResource)
}
