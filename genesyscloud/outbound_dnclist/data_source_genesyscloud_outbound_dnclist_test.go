package outbound_dnclist

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceOutboundDncList(t *testing.T) {
	var (
		resourceLabel   = "dnc_list"
		dncListName     = "Test List " + uuid.NewString()
		dataSourceLabel = "dnc_list_data"
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateOutboundDncListBasic(
					resourceLabel,
					dncListName,
				) + generateOutboundDncListDataSource(
					dataSourceLabel,
					dncListName,
					"genesyscloud_outbound_dnclist."+resourceLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_outbound_dnclist."+dataSourceLabel, "id",
						"genesyscloud_outbound_dnclist."+resourceLabel, "id"),
				),
			},
		},
	})
}

func generateOutboundDncListDataSource(dataSourceLabel string, attemptLimitName string, dependsOn string) string {
	return fmt.Sprintf(`
data "genesyscloud_outbound_dnclist" "%s" {
	name       = "%s"
	depends_on = [%s]
}
`, dataSourceLabel, attemptLimitName, dependsOn)
}
