package conversations_messaging_integrations_open

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"

	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"

	cmMessagingSetting "terraform-provider-genesyscloud/genesyscloud/conversations_messaging_settings"
	cmSupportedContent "terraform-provider-genesyscloud/genesyscloud/conversations_messaging_supportedcontent"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

/*
The resource_genesyscloud_conversations_messaging_integrations_open_test.go contains all of the test cases for running the resource
tests for conversations_messaging_integrations_open.
*/

func TestAccResourceConversationsMessagingIntegrationsOpen(t *testing.T) {
	t.Parallel()
	var (
		resourceLabel                                   = "test_messaging_open"
		name                                            = "Terraform Integrations Messaging Open " + uuid.NewString()
		outboundNotificationWebhookUrl1                 = "https://mock-server.prv-use1.test-pure.cloud/messaging-service/webhook"
		outboundNotificationWebhookSignatureSecretToken = uuid.NewString()

		nameSupportedContent       = "Terraform Supported Content - " + uuid.NewString()
		resourceIdSupportedContent = "testSupportedContent"
		inboundType                = "*/*"

		nameMessagingSetting       = "testSettings"
		resourceIdMessagingSetting = "testConversationsMessagingSettings"
	)

	supportedContentResource1 := cmSupportedContent.GenerateSupportedContentResource(
		"genesyscloud_conversations_messaging_supportedcontent",
		resourceIdSupportedContent,
		nameSupportedContent,
		cmSupportedContent.GenerateInboundTypeBlock(inboundType))

	messagingSettingResource1 := cmMessagingSetting.GenerateConversationsMessagingSettingsResource(
		resourceIdMessagingSetting,
		nameMessagingSetting,
		cmMessagingSetting.GenerateContentStoryBlock(
			cmMessagingSetting.GenerateMentionInboundOnlySetting("Disabled"),
			cmMessagingSetting.GenerateReplyInboundOnlySetting("Enabled"),
		),
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			//create
			{
				Config: messagingSettingResource1 +
					supportedContentResource1 +
					GenerateConversationMessagingOpenResource(
						resourceLabel,
						name,
						"genesyscloud_conversations_messaging_supportedcontent."+resourceIdSupportedContent+".id",
						"genesyscloud_conversations_messaging_settings."+resourceIdMessagingSetting+".id",
						outboundNotificationWebhookUrl1,
						outboundNotificationWebhookSignatureSecretToken,
						GenerateWebhookHeadersProperties("key", "value"),
					),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_integrations_open."+resourceLabel, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_integrations_open."+resourceLabel, "outbound_notification_webhook_url", outboundNotificationWebhookUrl1),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_integrations_open."+resourceLabel, "outbound_notification_webhook_signature_secret_token", outboundNotificationWebhookSignatureSecretToken),
					resource.TestCheckResourceAttrPair("genesyscloud_conversations_messaging_integrations_open."+resourceLabel, "supported_content_id", "genesyscloud_conversations_messaging_supportedcontent."+resourceIdSupportedContent, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_conversations_messaging_integrations_open."+resourceLabel, "messaging_setting_id", "genesyscloud_conversations_messaging_settings."+resourceIdMessagingSetting, "id"),
				),
			},
			{
				// Import/Read
				ResourceName:            "genesyscloud_conversations_messaging_integrations_open." + resourceLabel,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"outbound_notification_webhook_signature_secret_token"},
			},
		},
		CheckDestroy: testVerifyConversationsMessagingIntegrationsOpenDestroyed,
	})
}

func testVerifyConversationsMessagingIntegrationsOpenDestroyed(state *terraform.State) error {
	integrationApi := platformclientv2.NewConversationsApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_conversations_messaging_integrations_open" {
			continue
		}

		messagingOpen, resp, err := integrationApi.GetConversationsMessagingIntegrationsOpenIntegrationId(rs.Primary.ID, "")
		if messagingOpen != nil {
			return fmt.Errorf("Integration Messaging Open still exists")
		} else if util.IsStatus404(resp) {
			continue
		} else {
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	return nil
}
