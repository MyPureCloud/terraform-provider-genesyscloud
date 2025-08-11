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
		resourceName                  = "Terraform Messaging Whatsapp-" + uuid.NewString()
		resourceLabelSupportedContent = "testSupportedContent"
		nameSupportedContent          = "Terraform SupportedContent-" + uuid.NewString()
		inboundType                   = "*/*"

		resourceLabelMessagingSetting = "testMessagingSetting"
		nameMessagingSetting          = "TestTerraformMessagingSetting-" + uuid.NewString()

		embeddedToken = uuid.NewString()
	)

	if cleanupErr := CleanupMessagingSettings(nameMessagingSetting); cleanupErr != nil {
		t.Logf("Failed to clean up messaging settings with name '%s': %s", nameMessagingSetting, cleanupErr.Error())
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

func CleanupMessagingSettings(name string) error {
	cmMessagingSettingApi := platformclientv2.NewConversationsApiWithConfig(sdkConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		cmMessagingSetting, _, getErr := cmMessagingSettingApi.GetConversationsMessagingSettings(pageSize, pageNum)
		if getErr != nil {
			return fmt.Errorf("failed to get page %v of messaging settings: %v", pageNum, getErr)
		}

		if cmMessagingSetting.Entities == nil || len(*cmMessagingSetting.Entities) == 0 {
			break
		}

		for _, setting := range *cmMessagingSetting.Entities {
			if setting.Name != nil && strings.HasPrefix(*setting.Name, name) {
				log.Println("HIT: ", *setting.Name)
			}
		}
	}
	return nil
}
