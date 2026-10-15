package recording_settings

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

/*
Test Class for the Recording Settings Resource.

Note: This is an organization-level singleton resource with no create or delete API operation.
Because the settings always exist in the org, there is no CheckDestroy assertion (delete is a no-op).
*/

func TestAccResourceRecordingSettings(t *testing.T) {
	var (
		resourceLabel = "recording_settings"

		maxStreams1  = "100"
		playbackTtl1 = "60"
		batchTtl1    = "30"

		maxStreams2  = "150"
		playbackTtl2 = "45"
		batchTtl2    = "15"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create / initial update
				Config: generateRecordingSettingsResource(
					resourceLabel,
					maxStreams1,
					util.TrueValue,
					playbackTtl1,
					batchTtl1,
					util.TrueValue,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_recording_settings."+resourceLabel, "max_simultaneous_streams", maxStreams1),
					resource.TestCheckResourceAttr("genesyscloud_recording_settings."+resourceLabel, "regional_recording_storage_enabled", util.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_recording_settings."+resourceLabel, "recording_playback_url_ttl", playbackTtl1),
					resource.TestCheckResourceAttr("genesyscloud_recording_settings."+resourceLabel, "recording_batch_download_url_ttl", batchTtl1),
					resource.TestCheckResourceAttr("genesyscloud_recording_settings."+resourceLabel, "stop_recording_when_only_external_participants", util.TrueValue),
					resource.TestCheckResourceAttrSet("genesyscloud_recording_settings."+resourceLabel, "max_configurable_screen_recording_streams"),
				),
			},
			{
				// Update - change every writable attribute, including the TTL fields
				Config: generateRecordingSettingsResource(
					resourceLabel,
					maxStreams2,
					util.FalseValue,
					playbackTtl2,
					batchTtl2,
					util.FalseValue,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_recording_settings."+resourceLabel, "max_simultaneous_streams", maxStreams2),
					resource.TestCheckResourceAttr("genesyscloud_recording_settings."+resourceLabel, "regional_recording_storage_enabled", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_recording_settings."+resourceLabel, "recording_playback_url_ttl", playbackTtl2),
					resource.TestCheckResourceAttr("genesyscloud_recording_settings."+resourceLabel, "recording_batch_download_url_ttl", batchTtl2),
					resource.TestCheckResourceAttr("genesyscloud_recording_settings."+resourceLabel, "stop_recording_when_only_external_participants", util.FalseValue),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_recording_settings." + resourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccResourceRecordingSettingsTtlValidation verifies that the ValidateFunc on the
// TTL attributes rejects values outside the allowed 2-60 range at plan time.
func TestAccResourceRecordingSettingsTtlValidation(t *testing.T) {
	var (
		resourceLabel = "recording_settings"
		validTtl      = "60"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// playback TTL below the minimum (2)
				Config: generateRecordingSettingsResource(
					resourceLabel,
					"100",
					util.TrueValue,
					"1",
					validTtl,
					util.TrueValue,
				),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`expected recording_playback_url_ttl to be in the range \(2 - 60\)`),
			},
			{
				// batch download TTL above the maximum (60)
				Config: generateRecordingSettingsResource(
					resourceLabel,
					"100",
					util.TrueValue,
					validTtl,
					"61",
					util.TrueValue,
				),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`expected recording_batch_download_url_ttl to be in the range \(2 - 60\)`),
			},
		},
	})
}

func generateRecordingSettingsResource(
	resourceLabel string,
	maxSimultaneousStreams string,
	regionalStorageEnabled string,
	playbackUrlTtl string,
	batchDownloadUrlTtl string,
	stopRecordingExternalOnly string,
) string {
	return fmt.Sprintf(`resource "genesyscloud_recording_settings" "%s" {
	max_simultaneous_streams                       = %s
	regional_recording_storage_enabled             = %s
	recording_playback_url_ttl                     = %s
	recording_batch_download_url_ttl               = %s
	stop_recording_when_only_external_participants = %s
}
`, resourceLabel, maxSimultaneousStreams, regionalStorageEnabled, playbackUrlTtl, batchDownloadUrlTtl, stopRecordingExternalOnly)
}
