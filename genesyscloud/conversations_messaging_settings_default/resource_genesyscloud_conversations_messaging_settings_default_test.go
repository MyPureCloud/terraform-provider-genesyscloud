package conversations_messaging_settings_default

import (
	"fmt"
	conversationsMessagingSettings "terraform-provider-genesyscloud/genesyscloud/conversations_messaging_settings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceConversationsMessagingSettingsDefault(t *testing.T) {
	t.Parallel()
	var (
		messagingSettingsResourceLabel = "testConversationsMessagingSettings"
		messagingSettingsName          = "testSettingsForDefault"

		defaultResourceLabel = "testConversationsMessagingSettingsDefault"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			{
				// Create Messaging Settings and Verify succesful creation
				Config: conversationsMessagingSettings.GenerateConversationsMessagingSettingsResource(
					messagingSettingsResourceLabel,
					messagingSettingsName,
					conversationsMessagingSettings.GenerateContentStoryBlock(
						conversationsMessagingSettings.GenerateMentionInboundOnlySetting("Enabled"),
						conversationsMessagingSettings.GenerateReplyInboundOnlySetting("Enabled"),
					),
					conversationsMessagingSettings.GenerateTypingOnSetting(
						"Enabled",
						"Enabled",
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_settings."+messagingSettingsResourceLabel, "name", messagingSettingsName),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_settings."+messagingSettingsResourceLabel, "content.0.story.0.mention.0.inbound", "Enabled"),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_settings."+messagingSettingsResourceLabel, "content.0.story.0.reply.0.inbound", "Enabled"),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_settings."+messagingSettingsResourceLabel, "event.0.typing.0.on.0.inbound", "Enabled"),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_settings."+messagingSettingsResourceLabel, "event.0.typing.0.on.0.outbound", "Enabled"),
				),
			},
			{
				// Generate the Conversations Messaging Settings Default Resource
				Config: conversationsMessagingSettings.GenerateConversationsMessagingSettingsResource(
					messagingSettingsResourceLabel,
					messagingSettingsName,
					conversationsMessagingSettings.GenerateContentStoryBlock(
						conversationsMessagingSettings.GenerateMentionInboundOnlySetting("Enabled"),
						conversationsMessagingSettings.GenerateReplyInboundOnlySetting("Enabled"),
					),
					conversationsMessagingSettings.GenerateTypingOnSetting(
						"Enabled",
						"Enabled",
					),
				) + generateConversationsMessagingSettingsDefaultResource(defaultResourceLabel, "genesyscloud_conversations_messaging_settings."+messagingSettingsResourceLabel+".id"),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_conversations_messaging_settings_default." + defaultResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyConversationsMessagingSettingsDefaultDestroyed,
	})
}

func testVerifyConversationsMessagingSettingsDefaultDestroyed(state *terraform.State) error {
	return nil
}

func generateConversationsMessagingSettingsDefaultResource(resourceLabel, settingId string) string {
	return fmt.Sprintf(`resource "genesyscloud_conversations_messaging_settings_default" "%s" {
		setting_id = %s
	}
	`, resourceLabel, settingId)
}
