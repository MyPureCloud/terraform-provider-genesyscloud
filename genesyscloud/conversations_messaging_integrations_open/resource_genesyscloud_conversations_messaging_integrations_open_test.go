package conversations_messaging_integrations_open

import (
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	cmMessagingSetting "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_settings"
	cmSupportedContent "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_supportedcontent"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

/*
The resource_genesyscloud_conversations_messaging_integrations_open_test.go contains all of the test cases for running the resource
tests for conversations_messaging_integrations_open.
*/

func TestAccResourceConversationsMessagingIntegrationsOpen(t *testing.T) {
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

	if cleanupErr := CleanupConversationsMessagingIntegrationsOpen("Terraform Integrations Messaging Open"); cleanupErr != nil {
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
				Check: resource.ComposeTestCheckFunc(
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

func CleanupConversationsMessagingIntegrationsOpen(name string) error {
	log.Printf("Cleaning up conversations messaging integrations open with name '%s'", name)
	conversationsApi := platformclientv2.NewConversationsApiWithConfig(sdkConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		openIntegrations, _, getErr := conversationsApi.GetConversationsMessagingIntegrationsOpen(pageSize, pageNum, "", "", "")
		if getErr != nil {
			return fmt.Errorf("failed to get page %v of open integrations: %v", pageNum, getErr)
		}

		if openIntegrations.Entities == nil || len(*openIntegrations.Entities) == 0 {
			break
		}

		for _, integration := range *openIntegrations.Entities {
			if integration.Name != nil && strings.HasPrefix(*integration.Name, name) {
				log.Printf("Deleting open integration: %v", *integration.Id)
				_, err := conversationsApi.DeleteConversationsMessagingIntegrationsOpenIntegrationId(*integration.Id)
				if err != nil {
					log.Printf("failed to delete open integration: %v", err)
					continue
				}
				log.Printf("Deleted open integration: %v", *integration.Id)
				time.Sleep(5 * time.Second)
			}
		}
	}
	log.Printf("Cleaned up conversations messaging integrations open with name '%s'", name)
	return nil
}
