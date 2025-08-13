package conversations_messaging_integrations_whatsapp

import (
	"fmt"
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
The resource_genesyscloud_conversations_messaging_integrations_whatsapp_test.go contains all of the test cases for running the resource
tests for conversations_messaging_integrations_whatsapp.
*/

func TestAccResourceConversationsMessagingIntegrationsWhatsapp(t *testing.T) {
	t.Skip("Skipping because it requires setting up a org as test account for the mocks to respond correctly.")
	var (
		resourceLabel                 = "test_messaging_whatsapp"
		resourceName                  = "TestTerraformMessagingWhatsapp-" + uuid.NewString()
		resourceName2                 = "TestTerraformMessagingWhatsapp2-" + uuid.NewString()
		resourceLabelSupportedContent = "testSupportedContent"
		nameSupportedContent          = "TestTerraformSupportedContent-" + uuid.NewString()
		inboundType                   = "*/*"

		resourceLabelMessagingSetting = "testMessagingSetting"
		nameMessagingSetting          = "TestTerraformMessagingSetting-" + uuid.NewString()
		phoneNumber                   = "+13172222222"
		pin                           = "0000"
		embeddedToken                 = uuid.NewString()
	)

	if cleanupErr := CleanupMessagingIntegrationsWhatsapp("TestTerraformMessagingWhatsapp"); cleanupErr != nil {
		t.Logf("Failed to clean up conversations messaging integrations whatsapp with name '%s': %s", resourceName, cleanupErr.Error())
	}

	if cleanupErr := CleanupMessagingIntegrationsWhatsapp("TestTerraformMessagingWhatsapp2"); cleanupErr != nil {
		t.Logf("Failed to clean up conversations messaging integrations whatsapp with name '%s': %s", resourceName2, cleanupErr.Error())
	}

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
					),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "name", resourceName),
					resource.TestCheckResourceAttrPair(ResourceType+"."+resourceLabel, "supported_content_id", cmSupportedContent.ResourceType+"."+resourceLabelSupportedContent, "id"),
					resource.TestCheckResourceAttrPair(ResourceType+"."+resourceLabel, "messaging_setting_id", cmMessagingSetting.ResourceType+"."+resourceLabelMessagingSetting, "id"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "embedded_signup_access_token", embeddedToken),
				),
			},
			//update name and activate
			{
				Config: messagingSettingReference +
					supportedContentReference +
					GenerateConversationsMessagingIntegrationsWhatsappResource(
						resourceLabel,
						resourceName2,
						cmSupportedContent.ResourceType+"."+resourceLabelSupportedContent+".id",
						cmMessagingSetting.ResourceType+"."+resourceLabelMessagingSetting+".id",
						embeddedToken,
						GenerateActivateConversationsMessagingIntegrationsWhatsappResource(
							phoneNumber,
							pin),
					),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "name", resourceName2),
					resource.TestCheckResourceAttrPair(ResourceType+"."+resourceLabel, "supported_content_id", cmSupportedContent.ResourceType+"."+resourceLabelSupportedContent, "id"),
					resource.TestCheckResourceAttrPair(ResourceType+"."+resourceLabel, "messaging_setting_id", cmMessagingSetting.ResourceType+"."+resourceLabelMessagingSetting, "id"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "embedded_signup_access_token", embeddedToken),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "phone_number", phoneNumber),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "pin", pin),
				),
			},
			{
				// Import/Read
				ResourceName:      ResourceType + "." + resourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						time.Sleep(30 * time.Second) // Wait for 30 seconds for proper deletion
						return nil
					},
				),
				ImportStateVerifyIgnore: []string{"embedded_signup_access_token", "messaging_setting_id"},
			},
		},
		CheckDestroy: testVerifyConversationsMessagingIntegrationsWhatsappDestroyed,
	})
}

func testVerifyConversationsMessagingIntegrationsWhatsappDestroyed(state *terraform.State) error {
	integrationApi := platformclientv2.NewConversationsApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != ResourceType {
			continue
		}

		messagingWhatsapp, resp, err := integrationApi.GetConversationsMessagingIntegrationsWhatsappIntegrationId(rs.Primary.ID, "")
		if messagingWhatsapp != nil {
			return fmt.Errorf("Integration Messaging Whatsapp still exists")
		} else if util.IsStatus404(resp) {
			continue
		} else {
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	return nil
}
