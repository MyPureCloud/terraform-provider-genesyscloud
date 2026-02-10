package greeting_group

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
	groupResource "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/group"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

func TestAccResourceGroupGreeting(t *testing.T) {

	var (
		resourceLabel = "greeting"
		name1         = "Test Greeting " + uuid.NewString()
		type1         = "VOICEMAIL"
		ownerType1    = "GROUP"
		audioTts1     = "This is a test greeting"

		name2      = "Test Greeting " + uuid.NewString()
		type2      = "VOICEMAIL"
		ownerType2 = "GROUP"
		audioTts2  = "This is an updated test greeting"

		randomizer         = uuid.NewString()
		groupName          = "TestGroup" + randomizer
		groupResourceLabel = "sample_group"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			{
				Config: groupResource.GenerateBasicGroupResource(
					groupResourceLabel,
					groupName,
				) + GenerateGreetingGroup(
					resourceLabel,
					name1,
					type1,
					ownerType1,
					"genesyscloud_group."+groupResourceLabel+".id",
					audioTts1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_greeting_group."+resourceLabel, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_greeting_group."+resourceLabel, "type", type1),
					resource.TestCheckResourceAttrSet("genesyscloud_greeting_group."+resourceLabel, "owner_type"),
					resource.TestCheckResourceAttr("genesyscloud_greeting_group."+resourceLabel, "audio_tts", audioTts1),
					resource.TestCheckResourceAttrSet("genesyscloud_greeting_group."+resourceLabel, "group_id"),
				),
			},
			{
				Config: groupResource.GenerateBasicGroupResource(
					groupResourceLabel,
					groupName,
				) + GenerateGreetingGroup(
					resourceLabel,
					name2,
					type2,
					ownerType2,
					"genesyscloud_group."+groupResourceLabel+".id",
					audioTts2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_greeting_group."+resourceLabel, "name", name2),
					resource.TestCheckResourceAttr("genesyscloud_greeting_group."+resourceLabel, "type", type2),
					resource.TestCheckResourceAttrSet("genesyscloud_greeting_group."+resourceLabel, "owner_type"),
					resource.TestCheckResourceAttr("genesyscloud_greeting_group."+resourceLabel, "audio_tts", audioTts2),
					resource.TestCheckResourceAttrSet("genesyscloud_greeting_group."+resourceLabel, "group_id"),
				),
			},
			{
				ResourceName:            "genesyscloud_greeting_group." + resourceLabel,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"group_id"},
			},
		},
		//CheckDestroy: testVerifyGreetingDestroyed,
	})
}

func testVerifyGreetingDestroyed(state *terraform.State) error {
	greetingAPI := platformclientv2.NewGreetingsApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_greeting_group" {
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
