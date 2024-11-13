package integration_facebook

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	cmMessagingSetting "terraform-provider-genesyscloud/genesyscloud/conversations_messaging_settings"
	cmSupportedContent "terraform-provider-genesyscloud/genesyscloud/conversations_messaging_supportedcontent"
)

/*
Test Class for the integration facebook Data Source
*/

func TestAccDataSourceIntegrationFacebook(t *testing.T) {
	t.Skip("Skipping because it requires setting up a org as test account for the mocks to respond correctly.")
	t.Parallel()
	var (
		testResourceLabel1 = "test_sample"
		testResourceLabel2 = "test_sample2"
		name1              = "test_sample"
		pageAccessToken1   = uuid.NewString()
		appId              = ""
		appSecret          = ""

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
					generateFacebookIntegrationResource(
						testResourceLabel1,
						name1,
						"genesyscloud_conversations_messaging_supportedcontent."+resourceLabelSupportedContent+".id",
						"genesyscloud_conversations_messaging_settings."+resourceLabelMessagingSetting+".id",
						pageAccessToken1,
						"",
						"",
						appId,
						appSecret,
					) + generateIntegrationFacebookDataSource(
					testResourceLabel2,
					name1,
					"genesyscloud_integration_facebook."+testResourceLabel1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_integration_facebook."+testResourceLabel2, "id", "genesyscloud_integration_facebook."+testResourceLabel1, "id"),
				),
			},
		},
	})
}

func generateIntegrationFacebookDataSource(
	resourceLabel string,
	name string,
	dependsOnResource string) string {
	return fmt.Sprintf(`
	data "genesyscloud_integration_facebook" "%s" {
		name = "%s"
		depends_on = [%s]
	}
	`, resourceLabel, name, dependsOnResource)
}
