package conversations_messaging_supportedcontent_default

import (
	"fmt"
	"log"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	supportedContent "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_supportedcontent"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

/*
The resource_genesyscloud_conversations_messaging_supportedcontent_default_test.go contains all of the test cases for running the resource
tests for conversations_messaging_supportedcontent_default.
*/

func TestAccResourceConversationsMessagingSupportedcontentDefault(t *testing.T) {
	t.Parallel()
	var (
		defaultResourceLabel = "testSupportedDefaultContent"

		name          = "TestTerraformSupportedContent-" + uuid.NewString()
		resourceLabel = "testSupportedContent"
		inboundType   = "*/*"
		outboundType  = "*/*"
	)

	cleanupErr := cleanUpOldSupportedContentDefaults()
	if cleanupErr != nil {
		t.Logf("Error cleaning up old supported content defaults: %v", cleanupErr)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			{
				Config: supportedContent.GenerateSupportedContentResource(
					"genesyscloud_conversations_messaging_supportedcontent",
					resourceLabel,
					name,
					supportedContent.GenerateInboundTypeBlock(inboundType),
					supportedContent.GenerateOutboundTypeBlock(outboundType),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_supportedcontent."+resourceLabel, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_supportedcontent."+resourceLabel, "media_types.0.allow.0.inbound.0.type", inboundType),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_supportedcontent."+resourceLabel, "media_types.0.allow.0.outbound.0.type", outboundType),
				),
			},
			{
				Config: supportedContent.GenerateSupportedContentResource(
					"genesyscloud_conversations_messaging_supportedcontent",
					resourceLabel,
					name,
					supportedContent.GenerateInboundTypeBlock(inboundType),
					supportedContent.GenerateOutboundTypeBlock(outboundType),
				) +
					GenerateSupportedContentDefaultResource(
						defaultResourceLabel,
						"genesyscloud_conversations_messaging_supportedcontent."+resourceLabel+".id",
					),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("genesyscloud_conversations_messaging_supportedcontent_default."+defaultResourceLabel, "content_id", "genesyscloud_conversations_messaging_supportedcontent."+resourceLabel, "id"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_conversations_messaging_supportedcontent_default." + defaultResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyConversationsMessagingSupportedcontentDefaultDestroyed,
	})
}

func GenerateSupportedContentDefaultResource(
	resourceLabel string,
	id string,
) string {
	return fmt.Sprintf(`
		resource "genesyscloud_conversations_messaging_supportedcontent_default" "%s" {
			content_id = %s
		}
	`, resourceLabel, id)
}

func testVerifyConversationsMessagingSupportedcontentDefaultDestroyed(state *terraform.State) error {
	return nil
}

func cleanUpOldSupportedContentDefaults() error {
	log.Printf("Cleaning up older conversations messaging supported content defaults")
	cmMessagingSettingApi := platformclientv2.NewConversationsApiWithConfig(sdkConfig)
	scDefault, _, err := cmMessagingSettingApi.GetConversationsMessagingSupportedcontentDefault()
	if err != nil || scDefault == nil {
		return fmt.Errorf("failed to get conversations messaging supported content default: %w", err)
	}

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		cmMessagingSetting, _, getErr := cmMessagingSettingApi.GetConversationsMessagingSupportedcontent(pageSize, pageNum)
		if getErr != nil {
			return fmt.Errorf("failed to get page %v of messaging settings: %w", pageNum, getErr)
		}

		if cmMessagingSetting.Entities == nil || len(*cmMessagingSetting.Entities) == 0 {
			break
		}
		for _, setting := range *cmMessagingSetting.Entities {
			if setting.Name != nil && (*setting.Name == "default") && (*setting.Id != *scDefault.Id) { // keep the current default only
				log.Printf("Deleting messaging settings: %v", *setting.Id)
				_, err := cmMessagingSettingApi.DeleteConversationsMessagingSupportedcontentSupportedContentId(*setting.Id)
				if err != nil {
					log.Printf("failed to delete messaging settings: %v", err)
					continue
				}
				log.Printf("Deleted messaging settings: %v", *setting.Id)
			}
		}
	}
	log.Printf("Cleaned up old conversations messaging supported content defaults")
	return nil
}
