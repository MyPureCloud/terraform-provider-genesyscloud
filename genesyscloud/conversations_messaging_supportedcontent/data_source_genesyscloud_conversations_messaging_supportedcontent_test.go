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
		resourceId   = "testSupportedContent"
		dataSourceId = "testSupportedContent_data"
		name         = "Terraform Supported Content - " + uuid.NewString()
		inboundType  = "*/*"
		outboundType = "image/*"
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
				) +
					GenerateDataSourceForSupportedContent(
						resourceName,
						dataSourceId,
						name,
						"genesyscloud_conversations_messaging_supportedcontent."+resourceId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_conversations_messaging_supportedcontent."+dataSourceId, "id", "genesyscloud_conversations_messaging_supportedcontent."+resourceId, "id"),
				),
			},
		},
	})
}

func GenerateDataSourceForSupportedContent(
	resourceName string,
	resourceId string,
	name string,
	dependsOn string,
) string {
	return fmt.Sprintf(`
	data "%s" "%s" {
		name = "%s"
		depends_on = [%s]
	}
	`, resourceName, resourceId, name, dependsOn)
}
