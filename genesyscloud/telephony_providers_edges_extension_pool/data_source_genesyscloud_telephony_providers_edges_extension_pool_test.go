package telephony_providers_edges_extension_pool

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceExtensionPoolBasic(t *testing.T) {
	t.Parallel()
	var (
		extensionPoolStartNumber = "2500"
		extensionPoolEndNumber   = "2599"
		extensionPoolRes         = "extensionPool"
		extensionPoolDataRes     = "extensionPoolData"
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateExtensionPoolResource(&ExtensionPoolStruct{
					extensionPoolRes,
					extensionPoolStartNumber,
					extensionPoolEndNumber,
					util.NullValue, // No description
				}) + generateExtensionPoolDataSource(extensionPoolDataRes,
					extensionPoolStartNumber,
					extensionPoolEndNumber,
					"genesyscloud_telephony_providers_edges_extension_pool."+extensionPoolRes),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_telephony_providers_edges_extension_pool."+extensionPoolDataRes, "id", "genesyscloud_telephony_providers_edges_extension_pool."+extensionPoolRes, "id"),
				),
			},
		},
	})
}

func generateExtensionPoolDataSource(
	resourceID string,
	startNumber string,
	endNumber string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_telephony_providers_edges_extension_pool" "%s" {
		start_number = "%s"
		end_number = "%s"
		depends_on=[%s]
	}
	`, resourceID, startNumber, endNumber, dependsOnResource)
}
