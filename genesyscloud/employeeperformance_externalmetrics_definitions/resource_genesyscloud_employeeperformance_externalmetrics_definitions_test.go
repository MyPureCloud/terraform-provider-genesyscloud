package employeeperformance_externalmetrics_definitions

import (
	"fmt"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func TestAccResourceEmployeePerformanceExternalMetricsDefintions(t *testing.T) {

	var (
		defintionResource = "external_metrics_definitions"
		name1             = "Defintion " + uuid.NewString()
		units             = []string{`Seconds`, `Percent`, `Number`, `Currency`}
		defaultTypes      = []string{`HigherIsBetter`, `LowerIsBetter`, `TargetArea`}

		name2 = "Defintion " + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateEmployeePerformanceExternalMetricsDefinitionsResource(
					defintionResource,
					name1,
					units[2],
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
						"enabled", util.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_employeeperformance_externalmetrics_definitions."+defintionResource,
						"unit", units[2]),
				),
			},
			{
				// Update
				Config: generateEmployeePerformanceExternalMetricsDefinitionsResource(
					defintionResource,
					name2,
					units[2],
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
						"enabled", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_employeeperformance_externalmetrics_definitions."+defintionResource,
						"unit", units[2]),
				),
			},
			{
				// Update unit
				Config: generateEmployeePerformanceExternalMetricsDefinitionsResource(
					defintionResource,
					name2,
					units[0],
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
						"enabled", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_employeeperformance_externalmetrics_definitions."+defintionResource,
						"unit", units[0]),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_employeeperformance_externalmetrics_definitions." + defintionResource,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyEmployeePerformanceExternalMetricsDefinitionsDestroyed,
	})
}

func generateEmployeePerformanceExternalMetricsDefinitionsResource(
	resourceId string,
	name string,
	unit string,
	precision string,
	defaultObjectiveType string,
	enabled string,
	additionalFields ...string,
) string {
	return fmt.Sprintf(`
		resource "genesyscloud_employeeperformance_externalmetrics_definitions" "%s"{
			name = "%s"
			unit = "%s"
			precision = %s
			default_objective_type = "%s"
			enabled = %s
			%s
		}
	`, resourceId, name, unit, precision, defaultObjectiveType, enabled, strings.Join(additionalFields, ","))
}

func testVerifyEmployeePerformanceExternalMetricsDefinitionsDestroyed(state *terraform.State) error {
	gamificationAPI := platformclientv2.NewGamificationApi()

	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_employeeperformance_externalmetrics_definition" {
			continue
		}

		definition, resp, err := gamificationAPI.GetEmployeeperformanceExternalmetricsDefinition(rs.Primary.ID)
		if definition != nil {
			return fmt.Errorf("Definition (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
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
