package conversations_messaging_integrations_whatsapp

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"
	"time"

	cmMessagingSetting "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_settings"
	cmSupportedContent "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_supportedcontent"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

/*
Test Class for the conversations messaging integrations whatsapp Data Source
*/

func TestAccDataSourceConversationsMessagingIntegrationsWhatsapp(t *testing.T) {
	t.Parallel()
	var (
		resourceLabel                 = "test_messaging_whatsapp"
		dataSourceLabel               = "data_messaging_whatsapp"
		resourceName                  = "Terraform Messaging Whatsapp-" + uuid.NewString()
		resourceLabelSupportedContent = "testSupportedContent"
		nameSupportedContent          = "Terraform SupportedContent-" + uuid.NewString()
		inboundType                   = "*/*"

		resourceLabelMessagingSetting = "testMessagingSetting"
		nameMessagingSetting          = "Terraform MessagingSetting-" + uuid.NewString()

		embeddedToken = uuid.NewString()
	)

	supportedContentReference := cmSupportedContent.GenerateSupportedContentResource(
		"genesyscloud_conversations_messaging_supportedcontent",
		resourceLabelSupportedContent,
		nameSupportedContent,
		cmSupportedContent.GenerateInboundTypeBlock(inboundType),
	)

	messagingSettingReference := cmMessagingSetting.GenerateConversationsMessagingSettingsResource(
		resourceLabelMessagingSetting,
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
				Config: messagingSettingReference +
					supportedContentReference +
					GenerateConversationsMessagingIntegrationsWhatsappResource(
						resourceLabel,
						resourceName,
						cmSupportedContent.ResourceType+"."+resourceLabelSupportedContent+".id",
						cmMessagingSetting.ResourceType+"."+resourceLabelMessagingSetting+".id",
						embeddedToken,
					) +
					GenerateConversationsMessagingIntegrationWhatsappDataSource(
						dataSourceLabel,
						resourceName,
						ResourceType+"."+resourceLabel,
					),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data."+ResourceType+"."+dataSourceLabel, "id", ResourceType+"."+resourceLabel, "id"),
					func(s *terraform.State) error {
						time.Sleep(45 * time.Second) // Wait for 45 seconds for resources to get deleted properly
						return nil
					},
				),
			},
		},
	})
}
