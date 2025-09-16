package conversations_messaging_integrations_open

import (
	"fmt"
	"testing"
	"time"

	cmMessagingSetting "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_settings"
	cmSupportedContent "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_supportedcontent"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

/*
Test Class for the conversations messaging integrations open Data Source
*/

func TestAccDataSourceConversationsMessagingIntegrationsOpen(t *testing.T) {
	var (
		dataSourceLabel                                 = "data_test_messaging_open"
		resourceLabel                                   = "test_messaging_open"
		name                                            = "IntegrationsOpen" + uuid.NewString()
		outboundNotificationWebhookUrl1                 = "https://mock-server.prv-use1.test-pure.cloud/messaging-service/webhook"
		outboundNotificationWebhookSignatureSecretToken = uuid.NewString()

		nameSupportedContent       = "Terraform Supported Content - " + uuid.NewString()
		resourceIdSupportedContent = "testSupportedContent"
		inboundType                = "*/*"

		nameMessagingSetting       = "testSettings"
		resourceIdMessagingSetting = "testConversationsMessagingSettings"
	)

	if cleanupErr := CleanupConversationsMessagingIntegrationsOpen("TestTerraformSupportedContent-" + uuid.NewString()); cleanupErr != nil {
		t.Logf("Failed to clean up conversations messaging integrations open with name '%s': %s", name, cleanupErr.Error())
	}

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
						resourceLabel,
						name,
						"genesyscloud_conversations_messaging_supportedcontent."+resourceIdSupportedContent+".id",
						"genesyscloud_conversations_messaging_settings."+resourceIdMessagingSetting+".id",
						outboundNotificationWebhookUrl1,
						outboundNotificationWebhookSignatureSecretToken,
						GenerateWebhookHeadersProperties("key", "value"),
					) + GenerateConversationsMessagingOpenDataSource(
					dataSourceLabel,
					name,
					"genesyscloud_conversations_messaging_integrations_open."+resourceLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_conversations_messaging_integrations_open."+dataSourceLabel, "id", "genesyscloud_conversations_messaging_integrations_open."+resourceLabel, "id"),
					func(s *terraform.State) error {
						time.Sleep(30 * time.Second) // Wait for 30 seconds for proper updation
						return nil
					},
				),
			},
		},
		CheckDestroy: func(state *terraform.State) error {
			time.Sleep(60 * time.Second)
			return testVerifyConversationsMessagingIntegrationsOpenDestroyed(state)
		},
	})
}

func GenerateConversationsMessagingOpenDataSource(
	resourceLabel string,
	name string,
	dependsOnResource string,
) string {
	return fmt.Sprintf(`
		data "genesyscloud_conversations_messaging_integrations_open" "%s" {
			name = "%s"
			depends_on = [%s]
		}
	`, resourceLabel, name, dependsOnResource)
}
