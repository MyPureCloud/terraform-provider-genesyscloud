package greeting

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	userResource "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/user"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

func TestAccResourceGreeting(t *testing.T) {

	var (
		resourceLabel = "greeting"
		name1         = "Test Greeting " + uuid.NewString()
		type1         = "STATION"
		ownerType1    = "USER"
		audioTts1     = "This is a test greeting"

		name2      = "Test Greeting " + uuid.NewString()
		type2      = "STATION"
		ownerType2 = "USER"
		audioTts2  = "This is an updated test greeting"

		randomizer        = uuid.NewString()
		userName          = "TestUser" + randomizer
		userEmail         = randomizer + "@website.com"
		userResourceLabel = "sample_user"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			{
				Config: userResource.GenerateBasicUserResource(
					userResourceLabel,
					userEmail,
					userName,
				) + GenerateGreeting(
					resourceLabel,
					name1,
					type1,
					ownerType1,
					"genesyscloud_user."+userResourceLabel+".id",
					audioTts1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_greeting."+resourceLabel, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_greeting."+resourceLabel, "type", type1),
					resource.TestCheckResourceAttrSet("genesyscloud_greeting."+resourceLabel, "owner_type"),
					resource.TestCheckResourceAttr("genesyscloud_greeting."+resourceLabel, "audio_tts", audioTts1),
					resource.TestCheckResourceAttrSet("genesyscloud_greeting."+resourceLabel, "owner_id"),
				),
			},
			{
				Config: userResource.GenerateBasicUserResource(
					userResourceLabel,
					userEmail,
					userName,
				) + GenerateGreeting(
					resourceLabel,
					name2,
					type2,
					ownerType2,
					"genesyscloud_user."+userResourceLabel+".id",
					audioTts2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_greeting."+resourceLabel, "name", name2),
					resource.TestCheckResourceAttr("genesyscloud_greeting."+resourceLabel, "type", type2),
					resource.TestCheckResourceAttrSet("genesyscloud_greeting."+resourceLabel, "owner_type"),
					resource.TestCheckResourceAttr("genesyscloud_greeting."+resourceLabel, "audio_tts", audioTts2),
					resource.TestCheckResourceAttrSet("genesyscloud_greeting."+resourceLabel, "owner_id"),
				),
			},
			{
				ResourceName:            "genesyscloud_greeting." + resourceLabel,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"owner_id"},
			},
		},
		CheckDestroy: testVerifyGreetingDestroyed,
	})
}

func testVerifyGreetingDestroyed(state *terraform.State) error {
	greetingAPI := platformclientv2.NewGreetingsApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_greeting" {
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
