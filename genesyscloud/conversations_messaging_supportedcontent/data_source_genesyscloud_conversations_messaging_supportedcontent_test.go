package conversations_messaging_supportedcontent

import (
	"fmt"
	"testing"

	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

/*
Test Class for the supported content Data Source
*/

func TestAccDataSourceSupportedContent(t *testing.T) {
	t.Parallel()
	var (
		resourceLabel   = "testSupportedContent"
		dataSourceLabel = "testSupportedContent_data"
		name            = "Terraform Supported Content - " + uuid.NewString()
		inboundType     = "*/*"
		outboundType    = "image/*"
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
				) +
					GenerateDataSourceForSupportedContent(
						ResourceType,
						dataSourceLabel,
						name,
						"genesyscloud_conversations_messaging_supportedcontent."+resourceLabel),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_conversations_messaging_supportedcontent."+dataSourceLabel, "id", "genesyscloud_conversations_messaging_supportedcontent."+resourceLabel, "id"),
				),
			},
		},
	})
}

func GenerateDataSourceForSupportedContent(
	resourceType string,
	resourceLabel string,
	name string,
	dependsOn string,
) string {
	return fmt.Sprintf(`
	data "%s" "%s" {
		name = "%s"
		depends_on = [%s]
	}
	`, resourceType, resourceLabel, name, dependsOn)
}
