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

func TestAccResourceConversationsMessagingSettings(t *testing.T) {
	var (
		resource1 = "testConversationsMessagingSettings"
		name1     = "testSettings"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create with Content Block
				Config: generateConversationsMessagingSettingsResource(
					resource1,
					name1,
					generateContentStoryBlock(
						generateMentionInboundOnlySetting("Disabled"),
						generateReplyInboundOnlySetting("Enabled"),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_settings."+resource1, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_settings."+resource1, "content.0.story.0.mention.0.inbound", "Disabled"),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_settings."+resource1, "content.0.story.0.reply.0.inbound", "Enabled"),
				),
			},
			{
				// Update and Add Event Block
				Config: generateConversationsMessagingSettingsResource(
					resource1,
					name1,
					generateContentStoryBlock(
						generateMentionInboundOnlySetting("Enabled"),
						generateReplyInboundOnlySetting("Enabled"),
					),
					generateTypingOnSetting(
						"Enabled",
						"Disabled",
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_settings."+resource1, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_settings."+resource1, "content.0.story.0.mention.0.inbound", "Enabled"),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_settings."+resource1, "content.0.story.0.reply.0.inbound", "Enabled"),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_settings."+resource1, "event.0.typing.0.on.0.inbound", "Enabled"),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_settings."+resource1, "event.0.typing.0.on.0.outbound", "Disabled"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_conversations_messaging_settings." + resource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifySettingDestroyed,
	})
}

func TestAccResourceConversationsMessagingSettingsContentOnly(t *testing.T) {
	var (
		resource2 = "testConversationsMessagingSettingsContentOnly"
		name2     = "testSettingsContentOnly"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateConversationsMessagingSettingsResource(
					resource2,
					name2,
					generateContentStoryBlock(
						generateMentionInboundOnlySetting("Disabled"),
						generateReplyInboundOnlySetting("Disabled"),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_settings."+resource2, "name", name2),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_settings."+resource2, "content.0.story.0.mention.0.inbound", "Disabled"),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_settings."+resource2, "content.0.story.0.reply.0.inbound", "Disabled"),
				),
			},
			{
				// Update
				Config: generateConversationsMessagingSettingsResource(
					resource2,
					name2,
					generateContentStoryBlock(
						generateMentionInboundOnlySetting("Enabled"),
						generateReplyInboundOnlySetting("Enabled"),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_settings."+resource2, "name", name2),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_settings."+resource2, "content.0.story.0.mention.0.inbound", "Enabled"),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_settings."+resource2, "content.0.story.0.reply.0.inbound", "Enabled"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_conversations_messaging_settings." + resource2,
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
		if rs.Type != "genesyscloud_conversations_messaging_settings" {
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
