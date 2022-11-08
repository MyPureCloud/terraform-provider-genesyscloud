package genesyscloud

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"strings"
	"testing"
)

func TestAccDataSourceOutboundSettings(t *testing.T) {
	var (
		settingsResource   = "test-settings"
		settingsDataSource = "test-settings-data"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Search by name
				Config: generateOutboundSettingsResource(
					settingsResource,
					"10",
					"0.5",
					"6.5",
					"ALL_CALLS",
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
				) + generateOutboundSettingsDataSource(
					settingsDataSource,
					"10",
					"0.5",
					"6.5",
					"ALL_CALLS",
					"genesyscloud_outbound_settings."+settingsResource,
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
					resource.TestCheckResourceAttrPair("data.genesyscloud_outbound_settings."+settingsDataSource, "max_calls_per_agent",
						"genesyscloud_outbound_settings."+settingsResource, "max_calls_per_agent"),
					resource.TestCheckResourceAttrPair("data.genesyscloud_outbound_settings."+settingsDataSource, "max_line_utilization",
						"genesyscloud_outbound_settings."+settingsResource, "max_line_utilization"),
					resource.TestCheckResourceAttrPair("data.genesyscloud_outbound_settings."+settingsDataSource, "abandon_seconds",
						"genesyscloud_outbound_settings."+settingsResource, "abandon_seconds"),
					resource.TestCheckResourceAttrPair("data.genesyscloud_outbound_settings."+settingsDataSource, "compliance_abandon_rate_denominator",
						"genesyscloud_outbound_settings."+settingsResource, "compliance_abandon_rate_denominator"),
					resource.TestCheckResourceAttrPair("data.genesyscloud_outbound_settings."+settingsDataSource, "automatic_time_zone_mapping.0.supported_countries.0",
						"genesyscloud_outbound_settings."+settingsResource, "automatic_time_zone_mapping.0.supported_countries.0"),
					resource.TestCheckResourceAttrPair("data.genesyscloud_outbound_settings."+settingsDataSource, "automatic_time_zone_mapping.0.callable_windows.0.mapped.0.earliest_callable_time",
						"genesyscloud_outbound_settings."+settingsResource, "automatic_time_zone_mapping.0.callable_windows.0.mapped.0.earliest_callable_time"),
					resource.TestCheckResourceAttrPair("data.genesyscloud_outbound_settings."+settingsDataSource, "automatic_time_zone_mapping.0.callable_windows.0.mapped.0.latest_callable_time",
						"genesyscloud_outbound_settings."+settingsResource, "automatic_time_zone_mapping.0.callable_windows.0.mapped.0.latest_callable_time"),
					resource.TestCheckResourceAttrPair("data.genesyscloud_outbound_settings."+settingsDataSource, "automatic_time_zone_mapping.0.callable_windows.0.unmapped.0.earliest_callable_time",
						"genesyscloud_outbound_settings."+settingsResource, "automatic_time_zone_mapping.0.callable_windows.0.unmapped.0.earliest_callable_time"),
					resource.TestCheckResourceAttrPair("data.genesyscloud_outbound_settings."+settingsDataSource, "automatic_time_zone_mapping.0.callable_windows.0.unmapped.0.latest_callable_time",
						"genesyscloud_outbound_settings."+settingsResource, "automatic_time_zone_mapping.0.callable_windows.0.unmapped.0.latest_callable_time"),
					resource.TestCheckResourceAttrPair("data.genesyscloud_outbound_settings."+settingsDataSource, "automatic_time_zone_mapping.0.callable_windows.0.unmapped.0.time_zone_id",
						"genesyscloud_outbound_settings."+settingsResource, "automatic_time_zone_mapping.0.callable_windows.0.unmapped.0.time_zone_id"),
				),
			},
		},
	})
}

func generateOutboundSettingsDataSource(
	resourceId string,
	maxCallsPerAgent string,
	maxLineUtilization string,
	abandonSeconds string,
	complianceAbandonRateDenominator string,
	dependsOnResource string,
	nestedBlocks ...string) string {
	return fmt.Sprintf(`
		data "genesyscloud_outbound_settings" "%s"{
			max_calls_per_agent = %s
  			max_line_utilization = %s
			abandon_seconds = %s
  			compliance_abandon_rate_denominator = "%s"
			%s
			depends_on=[%s]
		}
		`, resourceId, maxCallsPerAgent, maxLineUtilization, abandonSeconds, complianceAbandonRateDenominator, strings.Join(nestedBlocks, "\n"), dependsOnResource,
	)
}
