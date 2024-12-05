package conversations_messaging_supportedcontent_default

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"

	supportedContent "terraform-provider-genesyscloud/genesyscloud/conversations_messaging_supportedcontent"

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

		name          = "Terraform Supported Content - " + uuid.NewString()
		resourceLabel = "testSupportedContent"
		inboundType   = "*/*"
		outboundType  = "*/*"
	)

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
