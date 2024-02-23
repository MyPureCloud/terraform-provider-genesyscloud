package outbound_callabletimeset

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceOutboundCallableTimeset(t *testing.T) {
	var (
		resourceId         = "callable_timeset"
		dataSourceId       = "callable_timeset_data"
		callabeTimesetName = "Callable timeset " + uuid.NewString()
		timeZone           = "Africa/Abidjan"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateOutboundCallabletimeset(
					resourceId,
					callabeTimesetName,
					GenerateCallableTimesBlock(
						timeZone,
						GenerateTimeSlotsBlock("07:00:00", "18:00:00", "3"),
						GenerateTimeSlotsBlock("09:30:00", "22:30:00", "5"),
					),
				) + generateOutboundCallabletimesetDataSource(
					dataSourceId,
					callabeTimesetName,
					"genesyscloud_outbound_callabletimeset."+resourceId,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_outbound_callabletimeset."+dataSourceId, "id",
						"genesyscloud_outbound_callabletimeset."+resourceId, "id"),
				),
			},
		},
	})
}

func generateOutboundCallabletimesetDataSource(id string, name string, dependsOn string) string {
	return fmt.Sprintf(`
	data "genesyscloud_outbound_callabletimeset" "%s" {
		name = "%s"
		depends_on = [%s]
	}
`, id, name, dependsOn)
}
