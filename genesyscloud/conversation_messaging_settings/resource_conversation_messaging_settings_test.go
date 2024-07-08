package conversations_messaging_settings

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceConversationMessagingSettings(t *testing.T) {
	var (
		resource1 = "test_ConversationMessagingSettings"
		name1     = "testSettings"

		inbound = "Disabled"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateFullConversationMessagingSettingsResource(
					resource1,
					name1,
					inbound,
					inbound,
					inbound,
					"Enabled",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_settings"+resource1, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_settings"+resource1, "content.0.story.0.mention.0.inbound", "Disabled"),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_settings"+resource1, "content.0.story.0.reply.0.inbound", "Disabled"),
					resource.TestCheckResourceAttr("genesyscloud_conversation_messaging_settings"+resource1, "event.0.typing.0.on.0.inbound", "Disabled"),
					resource.TestCheckResourceAttr("genesyscloud_conversation_messaging_settings"+resource1, "event.0.typing.0.on.0.outbound", "Disabled"),
				),
			},
			{
				Config: GenerateFullConversationMessagingSettingsResource(
					resource1,
					name1,
					inbound,
					inbound,
					inbound,
					inbound,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_conversation_messaging_settings"+resource1, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_conversation_messaging_settings"+resource1, "content.0.story.0.mention.0.inbound", "Disabled"),
					resource.TestCheckResourceAttr("genesyscloud_conversation_messaging_settings"+resource1, "content.0.story.0.reply.0.inbound", "Disabled"),
					resource.TestCheckResourceAttr("genesyscloud_conversation_messaging_settings"+resource1, "event.0.typing.0.on.0.inbound", "Disabled"),
					resource.TestCheckResourceAttr("genesyscloud_conversation_messaging_settings"+resource1, "event.0.typing.0.on.0.outbound", "Disabled"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_conversation_messaging_settings." + resource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifySettingDestroyed,
	})
}

func testVerifySettingDestroyed(state *terraform.State) error {
	messagingAPI := platformclientv2.NewConversationsApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_conversation_messaging_settings" {
			continue
		}

		setting, resp, err := messagingAPI.GetConversationsMessagingSetting(rs.Primary.ID)
		if setting != nil {
			return fmt.Errorf("Messaging setting (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			continue
		} else {
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	return nil
}
