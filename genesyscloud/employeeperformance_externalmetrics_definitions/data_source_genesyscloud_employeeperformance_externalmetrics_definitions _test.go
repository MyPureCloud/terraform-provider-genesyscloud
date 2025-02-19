package employeeperformance_externalmetrics_definitions

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceEmployeePerformanceExternalMetricsDefinitions(t *testing.T) {
	t.Parallel()
	var (
		definitionResourceLabel = "defintion"
		definitionDataLabel     = "defintion_data"
		name                    = "Defintion " + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Search by name
				Config: generateEmployeePerformanceExternalMetricsDefinitionsResource(
					definitionResourceLabel,
					name,
					"Seconds",
					"1",
					"TargetArea",
					"true",
				) + generateEmployeePerformanceExternalMetricsDefinitionsDataSource(
					definitionDataLabel,
					name,
					"genesyscloud_employeeperformance_externalmetrics_definitions."+definitionResourceLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.genesyscloud_employeeperformance_externalmetrics_definitions."+definitionDataLabel, "id",
						"genesyscloud_employeeperformance_externalmetrics_definitions."+definitionResourceLabel, "id",
					),
				),
			},
		},
	})
}

func generateEmployeePerformanceExternalMetricsDefinitionsDataSource(resourceLabel string, name string, dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_employeeperformance_externalmetrics_definitions" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceLabel, name, dependsOnResource)
}
