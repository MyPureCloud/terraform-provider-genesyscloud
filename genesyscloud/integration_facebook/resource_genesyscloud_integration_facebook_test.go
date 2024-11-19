package integration_facebook

import (
	"fmt"
	cmMessagingSetting "terraform-provider-genesyscloud/genesyscloud/conversations_messaging_settings"
	cmSupportedContent "terraform-provider-genesyscloud/genesyscloud/conversations_messaging_supportedcontent"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

/*
The resource_genesyscloud_integration_facebook_test.go contains all of the test cases for running the resource
tests for integration_facebook.
*/

func TestAccResourceIntegrationFacebook(t *testing.T) {
	t.Skip("Skipping because it requires setting up a org as test account for the mocks to respond correctly.")
	t.Parallel()
	var (
		testResource1    = "test_sample"
		name1            = "Terraform Facebook1-" + uuid.NewString()
		pageAccessToken1 = uuid.NewString()
		userAccessToken1 = uuid.NewString()
		pageId           = "1"
		appId            = ""
		appSecret        = ""
		name2            = "Terraform Facebook2-" + uuid.NewString()

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
			//with only PageAccessToken
			{
				Config: messagingSettingResource1 +
					supportedContentResource1 +
					generateFacebookIntegrationResource(
						testResource1,
						name1,
						"genesyscloud_conversations_messaging_supportedcontent."+resourceIdSupportedContent+".id",
						"genesyscloud_conversations_messaging_settings."+resourceIdMessagingSetting+".id",
						pageAccessToken1,
						"",
						"",
						appId,
						appSecret,
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_integration_facebook."+testResource1, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_integration_facebook."+testResource1, "page_access_token", pageAccessToken1),
					resource.TestCheckResourceAttr("genesyscloud_integration_facebook."+testResource1, "user_access_token", ""),
					resource.TestCheckResourceAttr("genesyscloud_integration_facebook."+testResource1, "page_id", ""),
					resource.TestCheckResourceAttr("genesyscloud_integration_facebook."+testResource1, "app_id", appId),
					resource.TestCheckResourceAttr("genesyscloud_integration_facebook."+testResource1, "app_secret", appSecret),
					resource.TestCheckResourceAttrPair("genesyscloud_integration_facebook."+testResource1, "supported_content_id", "genesyscloud_conversations_messaging_supportedcontent."+resourceIdSupportedContent, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_integration_facebook."+testResource1, "messaging_setting_id", "genesyscloud_conversations_messaging_settings."+resourceIdMessagingSetting, "id"),
				),
			},
			// Update resource
			{
				Config: messagingSettingResource1 +
					supportedContentResource1 +
					generateFacebookIntegrationResource(
						testResource1,
						name2,
						"genesyscloud_conversations_messaging_supportedcontent."+resourceIdSupportedContent+".id",
						"genesyscloud_conversations_messaging_settings."+resourceIdMessagingSetting+".id",
						"",
						userAccessToken1,
						pageId,
						appId,
						appSecret,
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_integration_facebook."+testResource1, "name", name2),
					resource.TestCheckResourceAttr("genesyscloud_integration_facebook."+testResource1, "page_access_token", ""),
					resource.TestCheckResourceAttr("genesyscloud_integration_facebook."+testResource1, "user_access_token", userAccessToken1),
					resource.TestCheckResourceAttr("genesyscloud_integration_facebook."+testResource1, "page_id", pageId),
					resource.TestCheckResourceAttr("genesyscloud_integration_facebook."+testResource1, "app_id", appId),
					resource.TestCheckResourceAttr("genesyscloud_integration_facebook."+testResource1, "app_secret", appSecret),
					resource.TestCheckResourceAttrPair("genesyscloud_integration_facebook."+testResource1, "supported_content_id", "genesyscloud_conversations_messaging_supportedcontent."+resourceIdSupportedContent, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_integration_facebook."+testResource1, "messaging_setting_id", "genesyscloud_conversations_messaging_settings."+resourceIdMessagingSetting, "id"),
				),
			},
			// With UserAccessToken and PageId
			{
				Config: messagingSettingResource1 +
					supportedContentResource1 +
					generateFacebookIntegrationResource(
						testResource1,
						name1,
						"genesyscloud_conversations_messaging_supportedcontent."+resourceIdSupportedContent+".id",
						"genesyscloud_conversations_messaging_settings."+resourceIdMessagingSetting+".id",
						"",
						userAccessToken1,
						pageId,
						appId,
						appSecret,
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_integration_facebook."+testResource1, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_integration_facebook."+testResource1, "page_access_token", ""),
					resource.TestCheckResourceAttr("genesyscloud_integration_facebook."+testResource1, "user_access_token", userAccessToken1),
					resource.TestCheckResourceAttr("genesyscloud_integration_facebook."+testResource1, "page_id", pageId),
					resource.TestCheckResourceAttr("genesyscloud_integration_facebook."+testResource1, "app_id", appId),
					resource.TestCheckResourceAttr("genesyscloud_integration_facebook."+testResource1, "app_secret", appSecret),
					resource.TestCheckResourceAttrPair("genesyscloud_integration_facebook."+testResource1, "supported_content_id", "genesyscloud_conversations_messaging_supportedcontent."+resourceIdSupportedContent, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_integration_facebook."+testResource1, "messaging_setting_id", "genesyscloud_conversations_messaging_settings."+resourceIdMessagingSetting, "id"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_integration_facebook." + testResource1,
				ImportState:       true,
				ImportStateVerify: true,
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						time.Sleep(30 * time.Second) // Wait for 30 seconds for proper deletion
						return nil
					},
				),
			},
		},
		CheckDestroy: testVerifyIntegrationFacebookDestroyed,
	})
}

func testVerifyIntegrationFacebookDestroyed(state *terraform.State) error {
	integrationAPI := platformclientv2.NewConversationsApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_integration_facebook" {
			continue
		}

		facebook, resp, err := integrationAPI.GetConversationsMessagingIntegrationsFacebookIntegrationId(rs.Primary.ID, "")
		if facebook != nil {
			return fmt.Errorf("Facebook still exists")
		} else if util.IsStatus404(resp) {
			// Facebook not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. Facebook config destroyed
	return nil
}
