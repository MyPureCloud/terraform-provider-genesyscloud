package genesyscloud

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/ronanwatkins/terraform-plugin-sdk/v2/helper/resource"
	"github.com/ronanwatkins/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v56/platformclientv2"
)

func TestAccResourceRoutingLanguageBasic(t *testing.T) {
	var (
		langResource1 = "test-lang1"
		langName1     = "Terraform Lang" + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateRoutingLanguageResource(
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
		CheckDestroy: testVerifyLanguagesDestroyed,
	})
}

func generateRoutingLanguageResource(
	resourceID string,
	name string) string {
	return fmt.Sprintf(`resource "genesyscloud_routing_language" "%s" {
		name = "%s"
	}
	`, resourceID, name)
}

func testVerifyLanguagesDestroyed(state *terraform.State) error {
	languagesAPI := platformclientv2.NewLanguagesApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_routing_language" {
			continue
		}

		lang, resp, err := languagesAPI.GetRoutingLanguage(rs.Primary.ID)
		if lang != nil {
			if lang.State != nil && *lang.State == "deleted" {
				// Language deleted
				continue
			}
			return fmt.Errorf("Language (%s) still exists", rs.Primary.ID)
		} else if isStatus404(resp) {
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
