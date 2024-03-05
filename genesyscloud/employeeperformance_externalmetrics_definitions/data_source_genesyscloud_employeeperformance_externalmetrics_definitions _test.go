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
		defintionRes  = "defintion"
		defintionData = "defintion_data"
		name          = "Defintion " + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Search by name
				Config: generateEmployeePerformanceExternalMetricsDefinitionsResource(
					defintionRes,
					name,
					"Seconds",
					"1",
					"TargetArea",
					"true",
				) + generateEmployeePerformanceExternalMetricsDefinitionsDataSource(
					defintionData,
					name,
					"genesyscloud_employeeperformance_externalmetrics_definitions."+defintionRes,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.genesyscloud_employeeperformance_externalmetrics_definitions."+defintionData, "id",
						"genesyscloud_employeeperformance_externalmetrics_definitions."+defintionRes, "id",
					),
				),
			},
		},
	})
}

func generateEmployeePerformanceExternalMetricsDefinitionsDataSource(resourceID string, name string, dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_employeeperformance_externalmetrics_definitions" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceID, name, dependsOnResource)
}
