package telephony_providers_edges_extension_pool

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
)

func TestAccDataSourceExtensionPoolBasic(t *testing.T) {
	t.Parallel()
	var (
		extensionPoolStartNumber = "2500"
		extensionPoolEndNumber   = "2599"

		extensionPoolResourceLabel    = "extensionPool"
		extensionPoolResourceFullPath = ResourceType + "." + extensionPoolResourceLabel

		extensionPoolDataSourceLabel    = "extensionPoolData"
		extensionPoolDataSourceFullPath = fmt.Sprintf("data.%s.%s", ResourceType, extensionPoolDataSourceLabel)
	)

	cleanupExtensionPool(t, extensionPoolStartNumber, extensionPoolEndNumber)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateExtensionPoolResource(&ExtensionPoolStruct{
					extensionPoolResourceLabel,
					extensionPoolStartNumber,
					extensionPoolEndNumber,
					util.NullValue, // No description
				}) + generateExtensionPoolDataSource(extensionPoolDataSourceLabel,
					extensionPoolStartNumber,
					extensionPoolEndNumber,
					extensionPoolResourceFullPath),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(extensionPoolDataSourceFullPath, "id", extensionPoolResourceFullPath, "id"),
				),
			},
		},
	})
}

func generateExtensionPoolDataSource(resourceLabel, startNumber, endNumber, dependsOnResource string) string {
	return fmt.Sprintf(`data "%s" "%s" {
		start_number = "%s"
		end_number = "%s"
		depends_on=[%s]
	}
	`, ResourceType, resourceLabel, startNumber, endNumber, dependsOnResource)
}
