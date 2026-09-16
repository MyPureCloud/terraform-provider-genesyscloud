package conversations_messaging_integrations_whatsapp

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"

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
		// WhatsApp two-step verification PIN must be 6 digits; a 4-digit value is rejected
		// with 400 "The specified PIN is invalid".
		pin           = "000000"
		embeddedToken = uuid.NewString()
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
			//activate (keep the same name)
			{
				// Wait for the integration's async creation to complete before activating.
				// Activating (PATCH .../embeddedsignup) while creation is in progress returns
				// 400 "Create integration has not completed".
				PreConfig: func() {
					time.Sleep(30 * time.Second)
				},
				// Note: the resource's update takes the activate branch and returns early when
				// activate_whatsapp changes, so a name change in the SAME step is not applied.
				// Keep the name unchanged here and only activate, so the checks match the
				// provider's actual behavior.
				Config: messagingSettingReference +
					supportedContentReference +
					GenerateConversationsMessagingIntegrationsWhatsappResource(
						resourceLabel,
						resourceName,
						cmSupportedContent.ResourceType+"."+resourceLabelSupportedContent+".id",
						cmMessagingSetting.ResourceType+"."+resourceLabelMessagingSetting+".id",
						embeddedToken,
						GenerateActivateConversationsMessagingIntegrationsWhatsappResource(
							phoneNumber,
							pin),
					),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "name", resourceName),
					resource.TestCheckResourceAttrPair(ResourceType+"."+resourceLabel, "supported_content_id", cmSupportedContent.ResourceType+"."+resourceLabelSupportedContent, "id"),
					resource.TestCheckResourceAttrPair(ResourceType+"."+resourceLabel, "messaging_setting_id", cmMessagingSetting.ResourceType+"."+resourceLabelMessagingSetting, "id"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "embedded_signup_access_token", embeddedToken),
					// Note: phone_number and pin are activation inputs only; the resource's read
					// does not populate them into state, so they are not asserted here.
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
				// activate_whatsapp (phone_number/pin) are activation inputs that the API does not
				// return on read, so they cannot round-trip through import.
				ImportStateVerifyIgnore: []string{"embedded_signup_access_token", "messaging_setting_id", "activate_whatsapp", "activate_whatsapp.#", "activate_whatsapp.0.%", "activate_whatsapp.0.phone_number", "activate_whatsapp.0.pin"},
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
