package outbound

import (
	"fmt"
	"testing"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

var falseValue = "false"

func TestAccDataSourceCallAnalysisResponseSet(t *testing.T) {
	var (
		resourceId      = "cars"
		responseSetName = "Test CAR " + uuid.NewString()
		dataSourceId    = "cars_data"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateOutboundCallAnalysisResponseSetResource(
					resourceId,
					responseSetName,
					falseValue,
					"",
				) + generateOutboundCallAnalysisResponseSetDataSource(
					dataSourceId,
					responseSetName,
					"genesyscloud_outbound_callanalysisresponseset."+resourceId,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_outbound_callanalysisresponseset."+dataSourceId, "id",
						"genesyscloud_outbound_callanalysisresponseset."+resourceId, "id"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func generateOutboundCallAnalysisResponseSetDataSource(id string, name string, dependsOn string) string {
	return fmt.Sprintf(`
data "genesyscloud_outbound_callanalysisresponseset" "%s" {
	name = "%s"
	depends_on = [%s]
}
`, id, name, dependsOn)
}
