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
		resourceId   = "dnc_list"
		dncListName  = "Test List " + uuid.NewString()
		dataSourceId = "dnc_list_data"
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateOutboundDncListBasic(
					resourceId,
					dncListName,
				) + generateOutboundDncListDataSource(
					dataSourceId,
					dncListName,
					"genesyscloud_outbound_dnclist."+resourceId,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_outbound_dnclist."+dataSourceId, "id",
						"genesyscloud_outbound_dnclist."+resourceId, "id"),
				),
			},
		},
	})
}

func generateOutboundDncListDataSource(id string, attemptLimitName string, dependsOn string) string {
	return fmt.Sprintf(`
data "genesyscloud_outbound_dnclist" "%s" {
	name       = "%s"
	depends_on = [%s]
}
`, id, attemptLimitName, dependsOn)
}
