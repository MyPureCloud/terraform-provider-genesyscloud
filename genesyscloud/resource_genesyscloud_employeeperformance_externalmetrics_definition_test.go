package genesyscloud

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
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
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
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
		CheckDestroy: testVerifyEmployeePerformanceExternalMetricsDefinitionsDestroyed,
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

func testVerifyEmployeePerformanceExternalMetricsDefinitionsDestroyed(state *terraform.State) error {
	gamificationAPI := platformclientv2.NewGamificationApi()

	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_employeeperformance_externalmetrics_definitions" {
			continue
		}

		definition, resp, err := gamificationAPI.GetEmployeeperformanceExternalmetricsDefinition(rs.Primary.ID)
		if definition != nil {
			return fmt.Errorf("Definition (%s) still exists", rs.Primary.ID)
		} else if IsStatus404(resp) {
			// Definition not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}

	// Success. All definitions destroyed
	return nil
}
