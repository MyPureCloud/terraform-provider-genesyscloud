package outbound_settings

import (
	"fmt"
	"strconv"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceOutboundSettings(t *testing.T) {

	t.Parallel()
	var (
		resourceId                       = "outbound_settings"
		complianceAbandonRateDenominator = []string{"ALL_CALLS", "CALLS_THAT_REACHED_QUEUE"}
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			{
				// Update all non nested values
				Config: generateOutboundSettingsResource(
					resourceId,
					"5",
					"0.2",
					"12.6",
					complianceAbandonRateDenominator[1],
					util.FalseValue,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_settings."+resourceId, "max_calls_per_agent", "5"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_settings."+resourceId, "max_line_utilization", "0.2"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_settings."+resourceId, "reschedule_time_zone_skipped_contacts", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_settings."+resourceId, "abandon_seconds", "12.6"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_settings."+resourceId, "compliance_abandon_rate_denominator", complianceAbandonRateDenominator[1]),
				),
			},
			{
				// Update some non nested values and some nested values
				Config: generateOutboundSettingsResource(
					resourceId,
					"7",
					"0",
					"10.0",
					"",
					util.TrueValue,
					generateAutomaticTimeZoneMapping(
						[]string{"CA"},
						generateCallableWindowsBlock(
							generateMapped(
								"09:00",
								"",
							),
							generateUnmapped(
								"",
								"19:00",
								"",
							),
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_settings."+resourceId, "max_calls_per_agent", "7"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_settings."+resourceId, "abandon_seconds", "10"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_settings."+resourceId, "reschedule_time_zone_skipped_contacts", util.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_settings."+resourceId,
						"automatic_time_zone_mapping.0.supported_countries.0", "CA"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_settings."+resourceId,
						"automatic_time_zone_mapping.0.callable_windows.0.mapped.0.earliest_callable_time", "09:00"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_settings."+resourceId,
						"automatic_time_zone_mapping.0.callable_windows.0.unmapped.0.latest_callable_time", "19:00"),
				),
			},
			{
				// Update all values
				Config: generateOutboundSettingsResource(
					resourceId,
					"10",
					"0.5",
					"6.5",
					complianceAbandonRateDenominator[0],
					util.FalseValue,
					generateAutomaticTimeZoneMapping(
						[]string{"US"},
						generateCallableWindowsBlock(
							generateMapped(
								"08:00",
								"18:00",
							),
							generateUnmapped(
								"09:45",
								"20:30",
								"CET",
							),
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_settings."+resourceId, "max_calls_per_agent", "10"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_settings."+resourceId, "reschedule_time_zone_skipped_contacts", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_settings."+resourceId, "max_line_utilization", "0.5"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_settings."+resourceId, "abandon_seconds", "6.5"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_settings."+resourceId, "compliance_abandon_rate_denominator", complianceAbandonRateDenominator[0]),
					resource.TestCheckResourceAttr("genesyscloud_outbound_settings."+resourceId,
						"automatic_time_zone_mapping.0.supported_countries.0", "US"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_settings."+resourceId,
						"automatic_time_zone_mapping.0.callable_windows.0.mapped.0.earliest_callable_time", "08:00"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_settings."+resourceId,
						"automatic_time_zone_mapping.0.callable_windows.0.mapped.0.latest_callable_time", "18:00"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_settings."+resourceId,
						"automatic_time_zone_mapping.0.callable_windows.0.unmapped.0.earliest_callable_time", "09:45"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_settings."+resourceId,
						"automatic_time_zone_mapping.0.callable_windows.0.unmapped.0.latest_callable_time", "20:30"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_settings."+resourceId,
						"automatic_time_zone_mapping.0.callable_windows.0.unmapped.0.time_zone_id", "CET"),
				),
			},
			{
				// Import/Read
				ResourceName:            "genesyscloud_outbound_settings." + resourceId,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"max_calls_per_agent", "max_line_utilization", "abandon_seconds", "compliance_abandon_rate_denominator", "automatic_time_zone_mapping"},
			},
		},
	})
}

func generateOutboundSettingsResource(
	resourceId string,
	maxCallsPerAgent string,
	maxLineUtilization string,
	abandonSeconds string,
	complianceAbandonRateDenominator string,
	rescheduleTimeZoneSkippedContacts string,
	nestedBlocks ...string) string {
	return fmt.Sprintf(`
		resource "genesyscloud_outbound_settings" "%s"{
			max_calls_per_agent = %s
  			max_line_utilization = %s
			abandon_seconds = %s
  			compliance_abandon_rate_denominator = "%s"
			reschedule_time_zone_skipped_contacts = %s
			%s
		}
		`, resourceId, maxCallsPerAgent, maxLineUtilization, abandonSeconds, complianceAbandonRateDenominator, rescheduleTimeZoneSkippedContacts, strings.Join(nestedBlocks, "\n"),
	)
}

func generateAutomaticTimeZoneMapping(
	supportedCountries []string,
	attrs ...string) string {

	formattedCountries := ""
	for i, countries := range supportedCountries {
		if i > 0 {
			formattedCountries += ", "
		}
		formattedCountries += strconv.Quote(countries)
	}

	return fmt.Sprintf(`
		automatic_time_zone_mapping {
			supported_countries = [%s]
			%s
		}
		`, formattedCountries, strings.Join(attrs, "\n"),
	)
}

func generateCallableWindowsBlock(
	mapped string,
	unmapped string) string {
	return fmt.Sprintf(`
		callable_windows {
			%s
			%s
		}
		`, mapped, unmapped,
	)
}

func generateMapped(
	earliestCallableTime string,
	latestCallableTime string) string {
	return fmt.Sprintf(`
		mapped{
			earliest_callable_time = "%s"
			latest_callable_time = "%s"
		}
		`, earliestCallableTime, latestCallableTime,
	)
}

func generateUnmapped(
	earliestCallableTime string,
	latestCallableTime string,
	timeZoneId string) string {
	return fmt.Sprintf(`
		unmapped{
			earliest_callable_time = "%s"
			latest_callable_time = "%s"
			time_zone_id = "%s"
		}
		`, earliestCallableTime, latestCallableTime, timeZoneId,
	)
}
