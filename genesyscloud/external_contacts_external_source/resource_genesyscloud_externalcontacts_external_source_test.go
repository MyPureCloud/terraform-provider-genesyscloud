package external_contacts_external_source

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

func TestAccResourceExternalSources(t *testing.T) {
	var (
		resourceLabel = "external_source"
		resourcePath  = ResourceType + "." + resourceLabel
		name          = "external-source-" + uuid.NewString()
		active        = true
		uri_template  = "https://some.host/{{externalId.value}}"
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateBasicExternalSourceResource(
					resourceLabel,
					name,
					active,
					uri_template,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "name", name),
					resource.TestCheckResourceAttr(resourcePath, "active", strconv.FormatBool(active)),
					resource.TestCheckResourceAttr(resourcePath, "link_configuration.0.uri_template", uri_template),
				),
			},

			{
				//Update
				Config: GenerateBasicExternalSourceResource(
					resourceLabel,
					name,
					active,
					uri_template,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "name", name),
					resource.TestCheckResourceAttr(resourcePath, "active", strconv.FormatBool(active)),
					resource.TestCheckResourceAttr(resourcePath, "link_configuration.0.uri_template", uri_template),
				),
			},
			{
				// Import/Read
				ResourceName:      resourcePath,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifySourceDestroyed,
	})
}

func GenerateBasicExternalSourceResource(
	resourceLabel string,
	name string,
	active bool,
	uri_template string,
) string {
	return fmt.Sprintf(`resource "%s" "%s" {
        name = "%s"
        active = %v
        link_configuration {
          uri_template = "%s"
        }
    }
    `, ResourceType, resourceLabel, name, active, uri_template)
}

func testVerifySourceDestroyed(state *terraform.State) error {
	externalAPI := platformclientv2.NewExternalContactsApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != ResourceType {
			continue
		}

		externalSource, resp, err := externalAPI.GetExternalcontactsExternalsource(rs.Primary.ID)
		if externalSource != nil {
			return fmt.Errorf("external source (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// External Source not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("unexpected error: %s", err)
		}
	}
	// Success. All external sources destroyed
	return nil
}
