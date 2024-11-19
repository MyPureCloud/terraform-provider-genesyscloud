package conversations_messaging_integrations_open

import (
	"fmt"
	"testing"

	cmMessagingSetting "terraform-provider-genesyscloud/genesyscloud/conversations_messaging_settings"
	cmSupportedContent "terraform-provider-genesyscloud/genesyscloud/conversations_messaging_supportedcontent"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

/*
Test Class for the conversations messaging integrations open Data Source
*/

func TestAccDataSourceConversationsMessagingIntegrationsOpen(t *testing.T) {
	t.Parallel()
	var (
		dataSourceId                                    = "data_test_messaging_open"
		resourceId                                      = "test_messaging_open"
		name                                            = "Data Terraform Integrations Messaging Open"
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
			{
				Config: messagingSettingResource1 +
					supportedContentResource1 +
					GenerateConversationMessagingOpenResource(
						resourceId,
						name,
						"genesyscloud_conversations_messaging_supportedcontent."+resourceIdSupportedContent+".id",
						"genesyscloud_conversations_messaging_settings."+resourceIdMessagingSetting+".id",
						outboundNotificationWebhookUrl1,
						outboundNotificationWebhookSignatureSecretToken,
						GenerateWebhookHeadersProperties("key", "value"),
					) + GenerateConversationsMessagingOpenDataSource(
					dataSourceId,
					name,
					"genesyscloud_conversations_messaging_integrations_open."+resourceId,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_conversations_messaging_integrations_open."+dataSourceId, "id", "genesyscloud_conversations_messaging_integrations_open."+resourceId, "id"),
				),
			},
		},
	})
}

func GenerateConversationsMessagingOpenDataSource(
	resourceId string,
	name string,
	dependsOnResource string,
) string {
	return fmt.Sprintf(`
		data "genesyscloud_conversations_messaging_integrations_open" "%s" {
			name = "%s"
			depends_on = [%s]
		}
	`, resourceId, name, dependsOnResource)
}
