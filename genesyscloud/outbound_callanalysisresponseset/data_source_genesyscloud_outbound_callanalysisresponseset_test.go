package outbound_callanalysisresponseset

import (
	"fmt"

	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

var FalseValue = "false"

func TestAccDataSourceOutboundCallAnalysisResponseSet(t *testing.T) {
	var (
		resourceLabel   = "cars"
		responseSetName = "Test CAR " + uuid.NewString()
		dataSourceLabel = "cars_data"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateOutboundCallAnalysisResponseSetResource(
					resourceLabel,
					responseSetName,
					FalseValue,
					"",
				) + generateOutboundCallAnalysisResponseSetDataSource(
					dataSourceLabel,
					responseSetName,
					"genesyscloud_outbound_callanalysisresponseset."+resourceLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_outbound_callanalysisresponseset."+dataSourceLabel, "id",
						"genesyscloud_outbound_callanalysisresponseset."+resourceLabel, "id"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func generateOutboundCallAnalysisResponseSetDataSource(dataSourceLabel string, name string, dependsOn string) string {
	return fmt.Sprintf(`
data "genesyscloud_outbound_callanalysisresponseset" "%s" {
	name = "%s"
	depends_on = [%s]
}
`, dataSourceLabel, name, dependsOn)
}
