package conversations_messaging_supportedcontent

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"

	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"

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
		name          = "Terraform Supported Content - " + uuid.NewString()
		inboundType   = "*/*"
		outboundType  = "*/*"
		inboundType2  = "image/*"
		inboundType3  = "video/mpeg"
	)

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
