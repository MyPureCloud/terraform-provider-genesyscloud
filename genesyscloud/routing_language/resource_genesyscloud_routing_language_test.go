package routing_language

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func TestAccResourceRoutingLanguageBasic(t *testing.T) {
	var (
		langResource1 = "test-lang1"
		langName1     = "Terraform Lang" + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateRoutingLanguageResource(
					langResource1,
					langName1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_language."+langResource1, "name", langName1),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_routing_language." + langResource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: ValidateLanguagesDestroyed,
	})
}

func ValidateLanguagesDestroyed(state *terraform.State) error {
	routingApi := platformclientv2.NewRoutingApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_routing_language" {
			continue
		}

		lang, resp, err := routingApi.GetRoutingLanguage(rs.Primary.ID)
		if lang != nil {
			if lang.State != nil && *lang.State == "deleted" {
				// Language deleted
				continue
			}
			return fmt.Errorf("Language (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// Language not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All languages destroyed
	return nil
}
