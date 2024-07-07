package integration_facebook

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The resource_genesyscloud_integration_facebook_test.go contains all of the test cases for running the resource
tests for integration_facebook.
*/

func TestAccResourceIntegrationFacebook(t *testing.T) {
	t.Parallel()
	var (
		testResource1       = "test_sample"
		name1               = "Terraform Facebook1-" + uuid.NewString()
		pageAccessToken1    = uuid.NewString()
		supportedContentId1 = "019c37a7-ccb4-4966-b1d7-ddb20399f7ab"
		messagingSettingId1 = "2c4e3b8e-3c9f-45c9-82cd-4bb54c8f18f0"
		userAccessToken1    = uuid.NewString()
		pageId              = "1"
		appId               = ""
		appSecret           = ""
		name2               = "Terraform Facebook2-" + uuid.NewString()
		// supportedContentId2 = "6862cf85-af34-492c-b1cf-b4bc0d33695f"
		// messagingSettingId2 = "2c4e3b8e-3c9f-45c9-82cd-4bb54c8f18f0"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			//with only PageAccessToken
			{
				Config: generateFacebookIntegrationResource(
					testResource1,
					name1,
					supportedContentId1,
					messagingSettingId1,
					pageAccessToken1,
					"",
					"",
					appId,
					appSecret,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_integration_facebook."+testResource1, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_integration_facebook."+testResource1, "supported_content_id", supportedContentId1),
					resource.TestCheckResourceAttr("genesyscloud_integration_facebook."+testResource1, "messaging_setting_id", messagingSettingId1),
					resource.TestCheckResourceAttr("genesyscloud_integration_facebook."+testResource1, "page_access_token", pageAccessToken1),
					resource.TestCheckResourceAttr("genesyscloud_integration_facebook."+testResource1, "user_access_token", ""),
					resource.TestCheckResourceAttr("genesyscloud_integration_facebook."+testResource1, "page_id", ""),
					resource.TestCheckResourceAttr("genesyscloud_integration_facebook."+testResource1, "app_id", appId),
					resource.TestCheckResourceAttr("genesyscloud_integration_facebook."+testResource1, "app_secret", appSecret),
				),
			},
			// Update resource
			{
				Config: generateFacebookIntegrationResource(
					testResource1,
					name2,
					supportedContentId1,
					messagingSettingId1,
					"",
					userAccessToken1,
					pageId,
					appId,
					appSecret,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_integration_facebook."+testResource1, "name", name2),
					resource.TestCheckResourceAttr("genesyscloud_integration_facebook."+testResource1, "supported_content_id", supportedContentId1),
					resource.TestCheckResourceAttr("genesyscloud_integration_facebook."+testResource1, "messaging_setting_id", messagingSettingId1),
					resource.TestCheckResourceAttr("genesyscloud_integration_facebook."+testResource1, "page_access_token", ""),
					resource.TestCheckResourceAttr("genesyscloud_integration_facebook."+testResource1, "user_access_token", userAccessToken1),
					resource.TestCheckResourceAttr("genesyscloud_integration_facebook."+testResource1, "page_id", pageId),
					resource.TestCheckResourceAttr("genesyscloud_integration_facebook."+testResource1, "app_id", appId),
					resource.TestCheckResourceAttr("genesyscloud_integration_facebook."+testResource1, "app_secret", appSecret),
				),
			},
			// With UserAccessToken and PageId
			{
				Config: generateFacebookIntegrationResource(
					testResource1,
					name1,
					supportedContentId1,
					messagingSettingId1,
					"",
					userAccessToken1,
					pageId,
					appId,
					appSecret,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_integration_facebook."+testResource1, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_integration_facebook."+testResource1, "supported_content_id", supportedContentId1),
					resource.TestCheckResourceAttr("genesyscloud_integration_facebook."+testResource1, "messaging_setting_id", messagingSettingId1),
					resource.TestCheckResourceAttr("genesyscloud_integration_facebook."+testResource1, "page_access_token", ""),
					resource.TestCheckResourceAttr("genesyscloud_integration_facebook."+testResource1, "user_access_token", userAccessToken1),
					resource.TestCheckResourceAttr("genesyscloud_integration_facebook."+testResource1, "page_id", pageId),
					resource.TestCheckResourceAttr("genesyscloud_integration_facebook."+testResource1, "app_id", appId),
					resource.TestCheckResourceAttr("genesyscloud_integration_facebook."+testResource1, "app_secret", appSecret),
				),
			},
			{
				// Import/Read
				ResourceName: "genesyscloud_integration_facebook." + testResource1,
				ImportState:  true,
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
