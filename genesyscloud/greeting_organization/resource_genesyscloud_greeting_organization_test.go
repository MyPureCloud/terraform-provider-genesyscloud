package greeting_organization

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

func TestAccResourceOrganizationGreeting(t *testing.T) {

	var (
		resourceLabel = "greeting"
		name1         = "Test Greeting " + uuid.NewString()
		type1         = "NAME"
		ownerType1    = "ORGANIZATION"
		audioTts1     = "This is a test greeting"

		name2      = "Test Greeting " + uuid.NewString()
		type2      = "NAME"
		ownerType2 = "ORGANIZATION"
		audioTts2  = "This is an updated test greeting"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			{
				Config: GenerateGreetingOrganization(
					resourceLabel,
					name1,
					type1,
					ownerType1,
					"",
					audioTts1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_greeting_organization."+resourceLabel, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_greeting_organization."+resourceLabel, "type", type1),
					resource.TestCheckResourceAttrSet("genesyscloud_greeting_organization."+resourceLabel, "owner_type"),
					resource.TestCheckResourceAttr("genesyscloud_greeting_organization."+resourceLabel, "audio_tts", audioTts1),
				),
			},
			{
				Config: GenerateGreetingOrganization(
					resourceLabel,
					name2,
					type2,
					ownerType2,
					"",
					audioTts2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_greeting_organization."+resourceLabel, "name", name2),
					resource.TestCheckResourceAttr("genesyscloud_greeting_organization."+resourceLabel, "type", type2),
					resource.TestCheckResourceAttrSet("genesyscloud_greeting_organization."+resourceLabel, "owner_type"),
					resource.TestCheckResourceAttr("genesyscloud_greeting_organization."+resourceLabel, "audio_tts", audioTts2),
				),
			},
			{
				ResourceName:      "genesyscloud_greeting_organization." + resourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyGreetingDestroyed,
	})
}

func testVerifyGreetingDestroyed(state *terraform.State) error {
	greetingAPI := platformclientv2.NewGreetingsApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_greeting_organization" {
			continue
		}
		greeting, resp, err := greetingAPI.GetGreeting(rs.Primary.ID)
		if greeting != nil {
			return fmt.Errorf("greeting (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			continue
		} else {
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	return nil
}
