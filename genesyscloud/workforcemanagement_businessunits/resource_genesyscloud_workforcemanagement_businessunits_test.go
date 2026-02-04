package workforcemanagement_businessunits

import (
	"fmt"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
)

func TestAccResourceWorkforcemanagementBusinessUnitBasic(t *testing.T) {
	var (
		buResourceLabel = "test-business-unit-settings"
		buName          = "TestBU" + uuid.NewString()
		buName2         = "TestBU2" + uuid.NewString()
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
					GenerateWorkforcemanagementBusinessUnitSettings(
						startDayOfWeek,
						timeZone,
						"",
						"",
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+buResourceLabel, "name", buName),
					resource.TestCheckResourceAttr(ResourceType+"."+buResourceLabel, "settings.0.start_day_of_week", startDayOfWeek),
					resource.TestCheckResourceAttr(ResourceType+"."+buResourceLabel, "settings.0.time_zone", timeZone),
				),
			},
			{
				// Update settings
				Config: GenerateWorkforcemanagementBusinessUnitResource(
					buResourceLabel,
					buName2,
					GenerateWorkforcemanagementBusinessUnitSettings(
						startDayOfWeek,
						timeZone,
						"",
						"",
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+buResourceLabel, "name", buName2),
					resource.TestCheckResourceAttr(ResourceType+"."+buResourceLabel, "settings.0.start_day_of_week", startDayOfWeek),
					resource.TestCheckResourceAttr(ResourceType+"."+buResourceLabel, "settings.0.time_zone", timeZone),
				),
			},
		},
		CheckDestroy: testVerifyBusinessUnitsDestroyed,
	})
}

func TestAccResourceWorkforcemanagementBusinessUnitWithShortTermForecasting(t *testing.T) {
	var (
		buResourceLabel = "test-business-unit-forecasting"
		buName          = "TestBU" + uuid.NewString()

		defaultHistoryWeeks  = "4"
		defaultHistoryWeeks2 = "8"
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
					GenerateWorkforcemanagementBusinessUnitSettings(
						"Monday",
						"America/New_York",
						GenerateWorkforcemanagementBusinessUnitShortTermForecasting(defaultHistoryWeeks),
						"",
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+buResourceLabel, "name", buName),
					resource.TestCheckResourceAttr(ResourceType+"."+buResourceLabel, "settings.0.short_term_forecasting.0.default_history_weeks", defaultHistoryWeeks),
				),
			},
			{
				// Update forecasting settings
				Config: GenerateWorkforcemanagementBusinessUnitResource(
					buResourceLabel,
					buName,
					GenerateWorkforcemanagementBusinessUnitSettings(
						"Monday",
						"America/New_York",
						GenerateWorkforcemanagementBusinessUnitShortTermForecasting(defaultHistoryWeeks2),
						"",
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+buResourceLabel, "settings.0.short_term_forecasting.0.default_history_weeks", defaultHistoryWeeks2),
				),
			},
		},
		CheckDestroy: testVerifyBusinessUnitsDestroyed,
	})
}

func TestAccResourceWorkforcemanagementBusinessUnitWithScheduling(t *testing.T) {
	var (
		buResourceLabel = "test-business-unit-scheduling"
		buName          = "TestBU" + uuid.NewString()

		messageType = "AgentWithoutCapability"
		severity    = "Warning"
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
					GenerateWorkforcemanagementBusinessUnitSettings(
						"Monday",
						"America/New_York",
						"",
						GenerateWorkforcemanagementBusinessUnitScheduling(
							"",
							nil,
							"",
							util.TrueValue,
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+buResourceLabel, "name", buName),
					resource.TestCheckResourceAttr(ResourceType+"."+buResourceLabel, "settings.0.scheduling.0.allow_work_plan_per_minute_granularity", util.TrueValue),
				),
			},
			{
				// Update with message severities
				Config: GenerateWorkforcemanagementBusinessUnitResource(
					buResourceLabel,
					buName,
					GenerateWorkforcemanagementBusinessUnitSettings(
						"Monday",
						"America/New_York",
						"",
						GenerateWorkforcemanagementBusinessUnitScheduling(
							GenerateWorkforcemanagementBusinessUnitMessageSeverities(messageType, severity),
							nil,
							"",
							util.FalseValue,
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+buResourceLabel, "settings.0.scheduling.0.message_severities.0.type", messageType),
					resource.TestCheckResourceAttr(ResourceType+"."+buResourceLabel, "settings.0.scheduling.0.message_severities.0.severity", severity),
					resource.TestCheckResourceAttr(ResourceType+"."+buResourceLabel, "settings.0.scheduling.0.allow_work_plan_per_minute_granularity", util.FalseValue),
				),
			},
		},
		CheckDestroy: testVerifyBusinessUnitsDestroyed,
	})
}

func TestAccResourceWorkforcemanagementBusinessUnitWithServiceGoalImpact(t *testing.T) {
	var (
		buResourceLabel = "test-business-unit-goal-impact"
		buName          = "TestBU" + uuid.NewString()

		serviceLevelIncreasePercent = "10"
		serviceLevelDecreasePercent = "5.5"

		averageSpeedOfAnswerIncreasePercent = "15"
		averageSpeedOfAnswerDecreasePercent = "10.7"

		abandonRateIncreasePercent = "20"
		abandonRateDecreasePercent = "15.1"
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
					GenerateWorkforcemanagementBusinessUnitSettings(
						"Monday",
						"America/New_York",
						"",
						GenerateWorkforcemanagementBusinessUnitScheduling(
							"",
							nil,
							GenerateWorkforcemanagementBusinessUnitServiceGoalImpact(
								GenerateWorkforcemanagementBusinessUnitServiceGoalImpactValue(serviceLevelIncreasePercent, serviceLevelDecreasePercent),
								GenerateWorkforcemanagementBusinessUnitServiceGoalImpactValue(averageSpeedOfAnswerIncreasePercent, averageSpeedOfAnswerDecreasePercent),
								GenerateWorkforcemanagementBusinessUnitServiceGoalImpactValue(abandonRateIncreasePercent, abandonRateDecreasePercent),
							),
							util.FalseValue,
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+buResourceLabel, "name", buName),
					resource.TestCheckResourceAttr(ResourceType+"."+buResourceLabel, "settings.0.scheduling.0.service_goal_impact.0.service_level.0.increase_by_percent", serviceLevelIncreasePercent),
					resource.TestCheckResourceAttr(ResourceType+"."+buResourceLabel, "settings.0.scheduling.0.service_goal_impact.0.service_level.0.decrease_by_percent", serviceLevelDecreasePercent),
					resource.TestCheckResourceAttr(ResourceType+"."+buResourceLabel, "settings.0.scheduling.0.service_goal_impact.0.average_speed_of_answer.0.increase_by_percent", averageSpeedOfAnswerIncreasePercent),
					resource.TestCheckResourceAttr(ResourceType+"."+buResourceLabel, "settings.0.scheduling.0.service_goal_impact.0.average_speed_of_answer.0.decrease_by_percent", averageSpeedOfAnswerDecreasePercent),
					resource.TestCheckResourceAttr(ResourceType+"."+buResourceLabel, "settings.0.scheduling.0.service_goal_impact.0.abandon_rate.0.increase_by_percent", abandonRateIncreasePercent),
					resource.TestCheckResourceAttr(ResourceType+"."+buResourceLabel, "settings.0.scheduling.0.service_goal_impact.0.abandon_rate.0.decrease_by_percent", abandonRateDecreasePercent),
				),
			},
		},
		CheckDestroy: testVerifyBusinessUnitsDestroyed,
	})
}

func testVerifyBusinessUnitsDestroyed(state *terraform.State) error {
	wfmAPI := platformclientv2.NewWorkforceManagementApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != ResourceType {
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

func generateHomeDivision() string {
	return fmt.Sprint(`data "genesyscloud_auth_division_home" "home" {}
`)
}
