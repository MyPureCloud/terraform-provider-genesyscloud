package workforcemanagement_businessunits

import (
	"fmt"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

func TestAccResourceWorkforcemanagementBusinessUnitBasic(t *testing.T) {
	var (
		buResourceLabel = "test-business-unit"
		buName          = "Terraform Test Business Unit " + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create basic business unit
				Config: GenerateWorkforcemanagementBusinessUnitResource(
					buResourceLabel,
					buName,
					util.NullValue, // Use default division
					"",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceName+"."+buResourceLabel, "name", buName),
				),
			},
			{
				// Import/Read
				ResourceName:      ResourceName + "." + buResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyBusinessUnitsDestroyed,
	})
}

func TestAccResourceWorkforcemanagementBusinessUnitWithSettings(t *testing.T) {
	var (
		buResourceLabel = "test-business-unit-settings"
		buName          = "Terraform Test BU Settings " + uuid.NewString()
		startDayOfWeek  = "Monday"
		timeZone        = "America/New_York"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create business unit with settings
				Config: GenerateWorkforcemanagementBusinessUnitResource(
					buResourceLabel,
					buName,
					util.NullValue,
					GenerateWorkforcemanagementBusinessUnitSettings(
						startDayOfWeek,
						timeZone,
						"",
						"",
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceName+"."+buResourceLabel, "name", buName),
					resource.TestCheckResourceAttr(ResourceName+"."+buResourceLabel, "settings.0.start_day_of_week", startDayOfWeek),
					resource.TestCheckResourceAttr(ResourceName+"."+buResourceLabel, "settings.0.time_zone", timeZone),
				),
			},
			{
				// Update settings
				Config: GenerateWorkforcemanagementBusinessUnitResource(
					buResourceLabel,
					buName,
					util.NullValue,
					GenerateWorkforcemanagementBusinessUnitSettings(
						"Sunday",
						"America/Los_Angeles",
						"",
						"",
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceName+"."+buResourceLabel, "settings.0.start_day_of_week", "Sunday"),
					resource.TestCheckResourceAttr(ResourceName+"."+buResourceLabel, "settings.0.time_zone", "America/Los_Angeles"),
				),
			},
		},
		CheckDestroy: testVerifyBusinessUnitsDestroyed,
	})
}

func TestAccResourceWorkforcemanagementBusinessUnitWithShortTermForecasting(t *testing.T) {
	var (
		buResourceLabel     = "test-business-unit-forecasting"
		buName              = "Terraform Test BU Forecasting " + uuid.NewString()
		defaultHistoryWeeks = 4
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create business unit with short term forecasting
				Config: GenerateWorkforcemanagementBusinessUnitResource(
					buResourceLabel,
					buName,
					util.NullValue,
					GenerateWorkforcemanagementBusinessUnitSettings(
						"Monday",
						"America/New_York",
						GenerateWorkforcemanagementBusinessUnitShortTermForecasting(defaultHistoryWeeks),
						"",
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceName+"."+buResourceLabel, "name", buName),
					resource.TestCheckResourceAttr(ResourceName+"."+buResourceLabel, "settings.0.short_term_forecasting.0.default_history_weeks", fmt.Sprintf("%d", defaultHistoryWeeks)),
				),
			},
			{
				// Update forecasting settings
				Config: GenerateWorkforcemanagementBusinessUnitResource(
					buResourceLabel,
					buName,
					util.NullValue,
					GenerateWorkforcemanagementBusinessUnitSettings(
						"Monday",
						"America/New_York",
						GenerateWorkforcemanagementBusinessUnitShortTermForecasting(8),
						"",
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceName+"."+buResourceLabel, "settings.0.short_term_forecasting.0.default_history_weeks", "8"),
				),
			},
		},
		CheckDestroy: testVerifyBusinessUnitsDestroyed,
	})
}

func TestAccResourceWorkforcemanagementBusinessUnitWithScheduling(t *testing.T) {
	var (
		buResourceLabel = "test-business-unit-scheduling"
		buName          = "Terraform Test BU Scheduling " + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create business unit with scheduling settings
				Config: GenerateWorkforcemanagementBusinessUnitResource(
					buResourceLabel,
					buName,
					util.NullValue,
					GenerateWorkforcemanagementBusinessUnitSettings(
						"Monday",
						"America/New_York",
						"",
						GenerateWorkforcemanagementBusinessUnitScheduling(
							"",
							"",
							"",
							util.TrueValue,
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceName+"."+buResourceLabel, "name", buName),
					resource.TestCheckResourceAttr(ResourceName+"."+buResourceLabel, "settings.0.scheduling.0.allow_work_plan_per_minute_granularity", "true"),
				),
			},
			{
				// Update with message severities
				Config: GenerateWorkforcemanagementBusinessUnitResource(
					buResourceLabel,
					buName,
					util.NullValue,
					GenerateWorkforcemanagementBusinessUnitSettings(
						"Monday",
						"America/New_York",
						"",
						GenerateWorkforcemanagementBusinessUnitScheduling(
							GenerateWorkforcemanagementBusinessUnitMessageSeverities("AgentSchedule", "Warning"),
							"",
							"",
							util.FalseValue,
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceName+"."+buResourceLabel, "settings.0.scheduling.0.message_severities.0.type", "AgentSchedule"),
					resource.TestCheckResourceAttr(ResourceName+"."+buResourceLabel, "settings.0.scheduling.0.message_severities.0.severity", "Warning"),
					resource.TestCheckResourceAttr(ResourceName+"."+buResourceLabel, "settings.0.scheduling.0.allow_work_plan_per_minute_granularity", "false"),
				),
			},
		},
		CheckDestroy: testVerifyBusinessUnitsDestroyed,
	})
}

func TestAccResourceWorkforcemanagementBusinessUnitWithServiceGoalImpact(t *testing.T) {
	var (
		buResourceLabel = "test-business-unit-goal-impact"
		buName          = "Terraform Test BU Goal Impact " + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create business unit with service goal impact settings
				Config: GenerateWorkforcemanagementBusinessUnitResource(
					buResourceLabel,
					buName,
					util.NullValue,
					GenerateWorkforcemanagementBusinessUnitSettings(
						"Monday",
						"America/New_York",
						"",
						GenerateWorkforcemanagementBusinessUnitScheduling(
							"",
							"",
							GenerateWorkforcemanagementBusinessUnitServiceGoalImpact(
								GenerateWorkforcemanagementBusinessUnitServiceGoalImpactValue(10.0, 5.0),
								GenerateWorkforcemanagementBusinessUnitServiceGoalImpactValue(15.0, 10.0),
								GenerateWorkforcemanagementBusinessUnitServiceGoalImpactValue(20.0, 15.0),
							),
							"",
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceName+"."+buResourceLabel, "name", buName),
					resource.TestCheckResourceAttr(ResourceName+"."+buResourceLabel, "settings.0.scheduling.0.service_goal_impact.0.service_level.0.increase_by_percent", "10"),
					resource.TestCheckResourceAttr(ResourceName+"."+buResourceLabel, "settings.0.scheduling.0.service_goal_impact.0.service_level.0.decrease_by_percent", "5"),
					resource.TestCheckResourceAttr(ResourceName+"."+buResourceLabel, "settings.0.scheduling.0.service_goal_impact.0.average_speed_of_answer.0.increase_by_percent", "15"),
					resource.TestCheckResourceAttr(ResourceName+"."+buResourceLabel, "settings.0.scheduling.0.service_goal_impact.0.abandon_rate.0.increase_by_percent", "20"),
				),
			},
		},
		CheckDestroy: testVerifyBusinessUnitsDestroyed,
	})
}

func testVerifyBusinessUnitsDestroyed(state *terraform.State) error {
	wfmAPI := platformclientv2.NewWorkforceManagementApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != ResourceName {
			continue
		}

		bu, resp, err := wfmAPI.GetWorkforcemanagementBusinessunit(rs.Primary.ID, []string{"settings"}, false)
		if bu != nil {
			return fmt.Errorf("Business unit (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// Business unit not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All business units destroyed
	return nil
}
