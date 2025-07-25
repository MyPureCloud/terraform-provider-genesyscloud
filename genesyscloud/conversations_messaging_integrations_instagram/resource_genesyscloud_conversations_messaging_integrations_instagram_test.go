package conversations_messaging_integrations_instagram

import (
	"fmt"
	"testing"

	cmMessagingSetting "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_settings"
	cmSupportedContent "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_supportedcontent"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

/*
The resource_genesyscloud_conversations_messaging_integrations_instagram_test.go contains all of the test cases for running the resource
tests for conversations_messaging_integrations_instagram.
*/

func TestAccResourceConversationsMessagingIntegrationsInstagram(t *testing.T) {
	t.Skip("Skipping because it requires setting up a org as test account for the mocks to respond correctly.")
	t.Parallel()

	var (
		testResourceLabel1 = "test_sample"
		name1              = "Terraform Instagram1-" + uuid.NewString()
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
			//create
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
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_integrations_instagram."+testResourceLabel1, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_integrations_instagram."+testResourceLabel1, "page_access_token", pageAccessToken1),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_integrations_instagram."+testResourceLabel1, "user_access_token", ""),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_integrations_instagram."+testResourceLabel1, "page_id", ""),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_integrations_instagram."+testResourceLabel1, "app_id", appId),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_integrations_instagram."+testResourceLabel1, "app_secret", appSecret),
					resource.TestCheckResourceAttrPair("genesyscloud_conversations_messaging_integrations_instagram."+testResourceLabel1, "supported_content_id", "genesyscloud_conversations_messaging_supportedcontent."+resourceLabelSupportedContent, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_conversations_messaging_integrations_instagram."+testResourceLabel1, "messaging_setting_id", "genesyscloud_conversations_messaging_settings."+resourceLabelMessagingSetting, "id"),
				),
			},
			// Test case commented as it requires setting up a org as test account for the updation/deletion mocks to respond correctly
			// {
			// 	//update
			// 	Config: messagingSettingResource1 +
			// 		supportedContentResource1 +
			// 		GenerateInstagramIntegrationResource(
			// 			testResource1,
			// 			name2,
			// 			"genesyscloud_conversations_messaging_supportedcontent."+resourceLabelSupportedContent+".id",
			// 			"genesyscloud_conversations_messaging_settings."+resourceLabelMessagingSetting+".id",
			// 			pageAccessToken1,
			// 			"",
			// 			"",
			// 			appId,
			// 			appSecret,
			// 		),

			// 	Check: resource.ComposeTestCheckFunc(
			// 		resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_integrations_instagram."+testResource1, "name", name2),
			// 		resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_integrations_instagram."+testResource1, "page_access_token", pageAccessToken1),
			// 		resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_integrations_instagram."+testResource1, "user_access_token", ""),
			// 		resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_integrations_instagram."+testResource1, "page_id", ""),
			// 		resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_integrations_instagram."+testResource1, "app_id", appId),
			// 		resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_integrations_instagram."+testResource1, "app_secret", appSecret),
			// 		resource.TestCheckResourceAttrPair("genesyscloud_conversations_messaging_integrations_instagram."+testResource1, "supported_content_id", "genesyscloud_conversations_messaging_supportedcontent."+resourceLabelSupportedContent, "id"),
			// 		resource.TestCheckResourceAttrPair("genesyscloud_conversations_messaging_integrations_instagram."+testResource1, "messaging_setting_id", "genesyscloud_conversations_messaging_settings."+resourceLabelMessagingSetting, "id"),
			// 	),
			// },
			// {
			// 	Config: messagingSettingResource1 +
			// 		supportedContentResource1 +
			// 		GenerateInstagramIntegrationResource(
			// 			testResource2,
			// 			name2,
			// 			"genesyscloud_conversations_messaging_supportedcontent."+resourceLabelSupportedContent+".id",
			// 			"genesyscloud_conversations_messaging_settings."+resourceLabelMessagingSetting+".id",
			// 			"",
			// 			userAccessToken1,
			// 			pageId,
			// 			appId,
			// 			appSecret,
			// 		),

			// 	Check: resource.ComposeTestCheckFunc(
			// 		resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_integrations_instagram."+testResource2, "name", name2),
			// 		resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_integrations_instagram."+testResource2, "page_access_token", ""),
			// 		resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_integrations_instagram."+testResource2, "user_access_token", userAccessToken1),
			// 		resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_integrations_instagram."+testResource2, "page_id", pageId),
			// 		resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_integrations_instagram."+testResource2, "app_id", appId),
			// 		resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_integrations_instagram."+testResource2, "app_secret", appSecret),
			// 		resource.TestCheckResourceAttrPair("genesyscloud_conversations_messaging_integrations_instagram."+testResource2, "supported_content_id", "genesyscloud_conversations_messaging_supportedcontent."+resourceLabelSupportedContent, "id"),
			// 		resource.TestCheckResourceAttrPair("genesyscloud_conversations_messaging_integrations_instagram."+testResource2, "messaging_setting_id", "genesyscloud_conversations_messaging_settings."+resourceLabelMessagingSetting, "id"),
			// 	),
			// },
			{
				// Import/Read
				ResourceName:            "genesyscloud_conversations_messaging_integrations_instagram." + testResourceLabel1,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"app_secret", "page_access_token", "user_access_token"},
			},
		},
		CheckDestroy: testVerifyConversationsMessagingIntegrationsInstagramDestroyed,
	})
}

func testVerifyConversationsMessagingIntegrationsInstagramDestroyed(state *terraform.State) error {
	integrationApi := platformclientv2.NewConversationsApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_conversations_messaging_integrations_instagram" {
			continue
		}

		instagram, resp, err := integrationApi.GetConversationsMessagingIntegrationsInstagramIntegrationId(rs.Primary.ID, "")
		if instagram != nil {
			return fmt.Errorf("Instagram still exists")
		} else if util.IsStatus404(resp) {
			continue
		} else {
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. Instagram config destroyed
	return nil
}
