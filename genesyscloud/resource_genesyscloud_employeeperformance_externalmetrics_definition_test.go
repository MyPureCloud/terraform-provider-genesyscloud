package genesyscloud

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccResourceEmployeePerformanceExternalMetricsDefintions(t *testing.T) {

	var (
		defintionResource = "external_metrics_definitions"
		name1             = "Defintion " + uuid.NewString()
		units             = []string{`Seconds`, `Percent`, `Number`, `Currency`}
		defaultTypes      = []string{`HigherIsBetter`, `LowerIsBetter`, `TargetArea`}
		description1      = "Example unit description"

		name2 = "Defintion " + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateEmployeePerformanceExternalMetricsDefinitionsResource(
					defintionResource,
					name1,
					units[2],
					nullValue,
					"5",
					defaultTypes[0],
					"true",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_employeeperformance_externalmetrics_definitions."+defintionResource,
						"name", name1),
					resource.TestCheckResourceAttr("genesyscloud_employeeperformance_externalmetrics_definitions."+defintionResource,
						"precision", "5"),
					resource.TestCheckResourceAttr("genesyscloud_employeeperformance_externalmetrics_definitions."+defintionResource,
						"default_objective_type", defaultTypes[0]),
					resource.TestCheckResourceAttr("genesyscloud_employeeperformance_externalmetrics_definitions."+defintionResource,
						"enabled", trueValue),
					resource.TestCheckResourceAttr("genesyscloud_employeeperformance_externalmetrics_definitions."+defintionResource,
						"unit", units[2]),
				),
			},
			{
				// Update
				Config: generateEmployeePerformanceExternalMetricsDefinitionsResource(
					defintionResource,
					name2,
					units[0],
					description1,
					"2",
					defaultTypes[1],
					"false",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_employeeperformance_externalmetrics_definitions."+defintionResource,
						"name", name2),
					resource.TestCheckResourceAttr("genesyscloud_employeeperformance_externalmetrics_definitions."+defintionResource,
						"precision", "2"),
					resource.TestCheckResourceAttr("genesyscloud_employeeperformance_externalmetrics_definitions."+defintionResource,
						"default_objective_type", defaultTypes[1]),
					resource.TestCheckResourceAttr("genesyscloud_employeeperformance_externalmetrics_definitions."+defintionResource,
						"enabled", falseValue),
					resource.TestCheckResourceAttr("genesyscloud_employeeperformance_externalmetrics_definitions."+defintionResource,
						"unit", units[0]),
					resource.TestCheckResourceAttr("genesyscloud_employeeperformance_externalmetrics_definitions."+defintionResource,
						"unit_definition", description1),
				),
			},
			{
				// Import/Read
				ResourceName:            "genesyscloud_employeeperformance_externalmetrics_definitions." + defintionResource,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"unit_definition"},
			},
		},
	})
}

func generateEmployeePerformanceExternalMetricsDefinitionsResource(
	resourceId string,
	name string,
	unit string,
	unitDefinition string,
	precision string,
	defaultObjectiveType string,
	enabled string,
) string {
	return fmt.Sprintf(`
		resource "genesyscloud_employeeperformance_externalmetrics_definitions" "%s"{
			name = "%s"
			unit = "%s"
			unit_definition = "%s"
			precision = %s
			default_objective_type = "%s"
			enabled = %s
		}
	`, resourceId, name, unit, unitDefinition, precision, defaultObjectiveType, enabled)
}
