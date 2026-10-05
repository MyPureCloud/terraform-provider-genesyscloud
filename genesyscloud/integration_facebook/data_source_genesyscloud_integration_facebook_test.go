package integration_facebook

import (
	"fmt"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	cmMessagingSetting "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_settings"
	cmSupportedContent "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_supportedcontent"
)

/*
Test Class for the integration facebook Data Source
*/

func TestAccDataSourceIntegrationFacebook(t *testing.T) {
	t.Parallel()
	var (
		testResourceLabel1 = "test_sample"
		testResourceLabel2 = "test_sample2"
		// Use a unique name so the data source's lookup-by-name matches the integration this
		// test creates, not a leftover "test_sample" integration from a prior run.
		name1            = "test_sample-" + uuid.NewString()
		pageAccessToken1 = uuid.NewString()
		appId            = ""
		appSecret        = ""

		nameSupportedContent          = "TestTerraformSupportedContent-" + uuid.NewString()
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
