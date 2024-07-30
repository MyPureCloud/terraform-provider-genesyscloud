package conversations_messaging_supportedcontent

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"

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
		resourceId   = "testSupportedContent"
		name         = "Terraform Supported Content - " + uuid.NewString()
		inboundType  = "*/*"
		outboundType = "*/*"
		inboundType2 = "image/*"
		inboundType3 = "video/mpeg"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateSupportedContentResource(
					resourceName,
					resourceId,
					name,
					GenerateInboundTypeBlock(inboundType),
					GenerateOutboundTypeBlock(outboundType),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_supportedcontent."+resourceId, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_supportedcontent."+resourceId, "media_types.0.allow.0.inbound.0.type", inboundType),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_supportedcontent."+resourceId, "media_types.0.allow.0.outbound.0.type", outboundType),
				),
			},
			//Update and add inbound block
			{
				Config: GenerateSupportedContentResource(
					resourceName,
					resourceId,
					name,
					GenerateInboundTypeBlock(inboundType2),
					GenerateInboundTypeBlock(inboundType3),
					GenerateOutboundTypeBlock(outboundType),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_supportedcontent."+resourceId, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_supportedcontent."+resourceId, "media_types.0.allow.0.inbound.0.type", inboundType2),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_supportedcontent."+resourceId, "media_types.0.allow.0.inbound.1.type", inboundType3),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_supportedcontent."+resourceId, "media_types.0.allow.0.outbound.0.type", outboundType),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_conversations_messaging_supportedcontent." + resourceId,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifySupportedContentDestroyed,
	})
}

func GenerateSupportedContentResource(
	resourceName string,
	resourceId string,
	name string,
	nestedBlocks ...string,
) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		name = "%s"
		media_types {
			allow {
				%s
			}
		}
	} `, resourceName, resourceId, name, strings.Join(nestedBlocks, "\n"))
}

func GenerateInboundTypeBlock(
	inboundType string,
) string {
	return fmt.Sprintf(`
		inbound {
			type="%s"
		}	
	`, inboundType)
}

func GenerateOutboundTypeBlock(
	outboundType string,
) string {
	return fmt.Sprintf(`
		outbound {
			type="%s"
		}	
	`, outboundType)
}

func testVerifySupportedContentDestroyed(state *terraform.State) error {
	supportContentApi := platformclientv2.NewConversationsApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != resourceName {
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
