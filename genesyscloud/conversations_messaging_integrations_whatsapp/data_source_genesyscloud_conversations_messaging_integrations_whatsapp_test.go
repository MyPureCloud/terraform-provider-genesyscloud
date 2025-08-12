package conversations_messaging_integrations_whatsapp

import (
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	cmMessagingSetting "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_settings"
	cmSupportedContent "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_supportedcontent"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

/*
Test Class for the conversations messaging integrations whatsapp Data Source
*/

func TestAccDataSourceConversationsMessagingIntegrationsWhatsapp(t *testing.T) {
	t.Parallel()
	var (
		resourceLabel                 = "test_messaging_whatsapp"
		dataSourceLabel               = "data_messaging_whatsapp"
		resourceName                  = "TestTerraformMessagingWhatsapp-" + uuid.NewString()
		resourceLabelSupportedContent = "testSupportedContent"
		nameSupportedContent          = "TestTerraformSupportedContent-" + uuid.NewString()
		inboundType                   = "*/*"

		resourceLabelMessagingSetting = "testMessagingSetting"
		nameMessagingSetting          = "TestTerraformMessagingSetting-" + uuid.NewString()

		embeddedToken = uuid.NewString()
	)

	if cleanupErr := CleanupMessagingIntegrationsWhatsapp("TestTerraformMessagingWhatsapp"); cleanupErr != nil {
		t.Logf("Failed to clean up conversations messaging integrations whatsapp with name '%s': %s", resourceName, cleanupErr.Error())
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

func CleanupMessagingIntegrationsWhatsapp(name string) error {
	whatsappApi := platformclientv2.NewConversationsApiWithConfig(sdkConfig)

	log.Printf("Cleaning up messaging integrations whatsapp with name '%s'", name)
	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		whatsappIntegrations, _, getErr := whatsappApi.GetConversationsMessagingIntegrationsWhatsapp(pageSize, pageNum, "", "", "")
		if getErr != nil {
			return fmt.Errorf("failed to get page %v of messaging settings: %v", pageNum, getErr)
		}

		if whatsappIntegrations.Entities == nil || len(*whatsappIntegrations.Entities) == 0 {
			break
		}

		for _, integration := range *whatsappIntegrations.Entities {
			if integration.Name != nil && strings.HasPrefix(*integration.Name, name) {
				_, resp, err := whatsappApi.DeleteConversationsMessagingIntegrationsWhatsappIntegrationId(*integration.Id)
				if err != nil {
					return fmt.Errorf("failed to delete messaging settings: %v | API Response: %s", err, resp)
				}
				time.Sleep(5 * time.Second)
			}
		}
	}

	log.Printf("Cleaned up messaging integrations whatsapp with name '%s'", name)
	return nil
}
