package recording_settings

import (
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

/*
Test Class for the Recording Settings Data Source.
*/

func TestAccDataSourceRecordingSettings(t *testing.T) {
	var (
		resourceLabel   = "recording_settings"
		dataSourceLabel = "recording_settings_data"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateRecordingSettingsResource(
					resourceLabel,
					"100",
					util.TrueValue,
					"60",
					"30",
					util.TrueValue,
				) + generateRecordingSettingsDataSource(dataSourceLabel, "genesyscloud_recording_settings."+resourceLabel),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.genesyscloud_recording_settings."+dataSourceLabel, "id",
						"genesyscloud_recording_settings."+resourceLabel, "id",
					),
					// The data source should expose the same values as the managed resource
					resource.TestCheckResourceAttrPair(
						"data.genesyscloud_recording_settings."+dataSourceLabel, "max_simultaneous_streams",
						"genesyscloud_recording_settings."+resourceLabel, "max_simultaneous_streams",
					),
					resource.TestCheckResourceAttrPair(
						"data.genesyscloud_recording_settings."+dataSourceLabel, "recording_playback_url_ttl",
						"genesyscloud_recording_settings."+resourceLabel, "recording_playback_url_ttl",
					),
					resource.TestCheckResourceAttrPair(
						"data.genesyscloud_recording_settings."+dataSourceLabel, "regional_recording_storage_enabled",
						"genesyscloud_recording_settings."+resourceLabel, "regional_recording_storage_enabled",
					),
				),
			},
		},
	})
}

func generateRecordingSettingsDataSource(dataSourceLabel string, dependsOnResource string) string {
	return `
data "genesyscloud_recording_settings" "` + dataSourceLabel + `" {
	depends_on = [` + dependsOnResource + `]
}
`
}
