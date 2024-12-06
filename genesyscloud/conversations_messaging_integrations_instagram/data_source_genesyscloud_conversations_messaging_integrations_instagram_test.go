package conversations_messaging_integrations_instagram

import (
	"testing"

	cmMessagingSetting "terraform-provider-genesyscloud/genesyscloud/conversations_messaging_settings"
	cmSupportedContent "terraform-provider-genesyscloud/genesyscloud/conversations_messaging_supportedcontent"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

/*
Test Class for the conversations messaging integrations instagram Data Source
*/

func TestAccDataSourceConversationsMessagingIntegrationsInstagram(t *testing.T) {
	t.Skip("Skipping because it requires setting up a org as test account for the mocks to respond correctly.")
	t.Parallel()
	var (
		testResourceLabel1  = "test_sample"
		testDataSourceLabel = "integration-instagram-ds"
		name1               = "Terraform Instagram1-" + uuid.NewString()
		pageAccessToken1    = uuid.NewString()
		appId               = ""
		appSecret           = ""

		nameSupportedContent          = "Terraform Supported Content - " + uuid.NewString()
		resourceLabelSupportedContent = "testSupportedContent"
		inboundType                   = "*/*"

		nameMessagingSetting          = "testSettings"
		resourceLabelMessagingSetting = "testConversationsMessagingSettings"
	)

	supportedContentResource1 := cmSupportedContent.GenerateSupportedContentResource(
		"genesyscloud_conversations_messaging_supportedcontent",
		resourceLabelSupportedContent,
		nameSupportedContent,
		cmSupportedContent.GenerateInboundTypeBlock(inboundType))

	messagingSettingResource1 := cmMessagingSetting.GenerateConversationsMessagingSettingsResource(
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
			{
				Config: messagingSettingResource1 +
					supportedContentResource1 +
					GenerateInstagramIntegrationResource(
						testResourceLabel1,
						name1,
						"genesyscloud_conversations_messaging_supportedcontent."+resourceLabelSupportedContent+".id",
						"genesyscloud_conversations_messaging_settings."+resourceLabelMessagingSetting+".id",
						pageAccessToken1,
						"",
						"",
						appId,
						appSecret,
					),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_conversations_messaging_integrations_instagram."+testDataSourceLabel, "id", "genesyscloud_conversations_messaging_integrations_instagram."+testResourceLabel1, "id"),
				),
			},
		},
	})
}
