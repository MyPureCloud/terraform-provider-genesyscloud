package conversations_messaging_supportedcontent

import (
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

/*
The resource_genesyscloud_conversations_messaging_supportedcontent_test.go contains all of the test cases for running the resource
tests for supported_content.
*/

func TestAccResourceSupportedContent(t *testing.T) {
	t.Parallel()
	var (
		resourceLabel = "testSupportedContent"
		name          = "TestTerraformSupportedContent-" + uuid.NewString()
		inboundType   = "*/*"
		outboundType  = "*/*"
		inboundType2  = "image/*"
		inboundType3  = "video/mpeg"
	)

	if cleanupErr := CleanupMessagingSettingsSupportedContent("TestTerraformSupportedContent-" + uuid.NewString()); cleanupErr != nil {
		t.Logf("Failed to clean up conversations messaging supported content with name '%s': %s", name, cleanupErr.Error())
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateSupportedContentResource(
					ResourceType,
					resourceLabel,
					name,
					GenerateInboundTypeBlock(inboundType),
					GenerateOutboundTypeBlock(outboundType),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_supportedcontent."+resourceLabel, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_supportedcontent."+resourceLabel, "media_types.0.allow.0.inbound.0.type", inboundType),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_supportedcontent."+resourceLabel, "media_types.0.allow.0.outbound.0.type", outboundType),
				),
			},
			//Update and add inbound block
			{
				Config: GenerateSupportedContentResource(
					ResourceType,
					resourceLabel,
					name,
					GenerateInboundTypeBlock(inboundType2),
					GenerateInboundTypeBlock(inboundType3),
					GenerateOutboundTypeBlock(outboundType),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_supportedcontent."+resourceLabel, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_supportedcontent."+resourceLabel, "media_types.0.allow.0.inbound.0.type", inboundType2),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_supportedcontent."+resourceLabel, "media_types.0.allow.0.inbound.1.type", inboundType3),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_supportedcontent."+resourceLabel, "media_types.0.allow.0.outbound.0.type", outboundType),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_conversations_messaging_supportedcontent." + resourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifySupportedContentDestroyed,
	})
}

func testVerifySupportedContentDestroyed(state *terraform.State) error {
	supportContentApi := platformclientv2.NewConversationsApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != ResourceType {
			continue
		}

		supportedContent, resp, err := supportContentApi.GetConversationsMessagingSupportedcontentSupportedContentId(rs.Primary.ID)

		if supportedContent != nil {
			return fmt.Errorf("Supported Content (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			continue
		} else {
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	return nil
}

func CleanupMessagingSettingsSupportedContent(name string) error {
	log.Printf("Cleaning up conversations messaging supported content with name '%s'", name)
	cmMessagingSettingApi := platformclientv2.NewConversationsApiWithConfig(sdkConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		cmMessagingSetting, _, getErr := cmMessagingSettingApi.GetConversationsMessagingSupportedcontent(pageSize, pageNum)
		if getErr != nil {
			return fmt.Errorf("failed to get page %v of messaging settings: %v", pageNum, getErr)
		}

		if cmMessagingSetting.Entities == nil || len(*cmMessagingSetting.Entities) == 0 {
			break
		}

		for _, setting := range *cmMessagingSetting.Entities {
			if setting.Name != nil && strings.HasPrefix(*setting.Name, name) {
				log.Printf("Deleting messaging settings: %v", *setting.Id)
				_, err := cmMessagingSettingApi.DeleteConversationsMessagingSupportedcontentSupportedContentId(*setting.Id)
				if err != nil {
					log.Printf("failed to delete messaging settings: %v", err)
					continue
				}
				log.Printf("Deleted messaging settings: %v", *setting.Id)
				time.Sleep(5 * time.Second)
			}
		}
	}
	log.Printf("Cleaned up conversations messaging supported content with name '%s'", name)
	return nil
}
