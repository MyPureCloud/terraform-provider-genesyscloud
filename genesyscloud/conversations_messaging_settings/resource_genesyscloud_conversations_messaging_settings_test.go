package conversations_messaging_settings

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v149/platformclientv2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceConversationsMessagingSettings(t *testing.T) {
	var (
		resourceLabel1 = "testConversationsMessagingSettings"
		name1          = "testSettings"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create with Content Block
				Config: GenerateConversationsMessagingSettingsResource(
					resourceLabel1,
					name1,
					GenerateContentStoryBlock(
						GenerateMentionInboundOnlySetting("Disabled"),
						GenerateReplyInboundOnlySetting("Enabled"),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_settings."+resourceLabel1, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_settings."+resourceLabel1, "content.0.story.0.mention.0.inbound", "Disabled"),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_settings."+resourceLabel1, "content.0.story.0.reply.0.inbound", "Enabled"),
				),
			},
			{
				// Update and Add Event Block
				Config: GenerateConversationsMessagingSettingsResource(
					resourceLabel1,
					name1,
					GenerateContentStoryBlock(
						GenerateMentionInboundOnlySetting("Enabled"),
						GenerateReplyInboundOnlySetting("Enabled"),
					),
					GenerateTypingOnSetting(
						"Enabled",
						"Disabled",
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_settings."+resourceLabel1, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_settings."+resourceLabel1, "content.0.story.0.mention.0.inbound", "Enabled"),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_settings."+resourceLabel1, "content.0.story.0.reply.0.inbound", "Enabled"),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_settings."+resourceLabel1, "event.0.typing.0.on.0.inbound", "Enabled"),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_settings."+resourceLabel1, "event.0.typing.0.on.0.outbound", "Disabled"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_conversations_messaging_settings." + resourceLabel1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifySettingDestroyed,
	})
}

func TestAccResourceConversationsMessagingSettingsContentOnly(t *testing.T) {
	var (
		resourceLabel2 = "testConversationsMessagingSettingsContentOnly"
		name2          = "testSettingsContentOnly"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateConversationsMessagingSettingsResource(
					resourceLabel2,
					name2,
					GenerateContentStoryBlock(
						GenerateMentionInboundOnlySetting("Disabled"),
						GenerateReplyInboundOnlySetting("Disabled"),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_settings."+resourceLabel2, "name", name2),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_settings."+resourceLabel2, "content.0.story.0.mention.0.inbound", "Disabled"),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_settings."+resourceLabel2, "content.0.story.0.reply.0.inbound", "Disabled"),
				),
			},
			{
				// Update
				Config: GenerateConversationsMessagingSettingsResource(
					resourceLabel2,
					name2,
					GenerateContentStoryBlock(
						GenerateMentionInboundOnlySetting("Enabled"),
						GenerateReplyInboundOnlySetting("Enabled"),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_settings."+resourceLabel2, "name", name2),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_settings."+resourceLabel2, "content.0.story.0.mention.0.inbound", "Enabled"),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_settings."+resourceLabel2, "content.0.story.0.reply.0.inbound", "Enabled"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_conversations_messaging_settings." + resourceLabel2,
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
