package recording_settings

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
)

/*
The resource_genesyscloud_recording_settings_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.
*/

// getRecordingSettingsFromResourceData maps the Terraform resource data into a platformclientv2.Recordingsettings struct.
//
// Note: max_configurable_screen_recording_streams is a read-only (Computed) field returned by the API and
// is intentionally not included in the update body.
func getRecordingSettingsFromResourceData(d *schema.ResourceData) platformclientv2.Recordingsettings {
	return platformclientv2.Recordingsettings{
		MaxSimultaneousStreams:                    resourcedata.GetNillableValue[int](d, "max_simultaneous_streams"),
		RegionalRecordingStorageEnabled:           resourcedata.GetNillableValue[bool](d, "regional_recording_storage_enabled"),
		RecordingPlaybackUrlTtl:                   resourcedata.GetNillableValue[int](d, "recording_playback_url_ttl"),
		RecordingBatchDownloadUrlTtl:              resourcedata.GetNillableValue[int](d, "recording_batch_download_url_ttl"),
		StopRecordingWhenOnlyExternalParticipants: resourcedata.GetNillableValue[bool](d, "stop_recording_when_only_external_participants"),
	}
}

// setRecordingSettingsToResourceData maps a platformclientv2.Recordingsettings struct onto the Terraform
// resource data. It is shared by both the resource read and the data source read so the mapping stays consistent.
func setRecordingSettingsToResourceData(d *schema.ResourceData, settings *platformclientv2.Recordingsettings) {
	resourcedata.SetNillableValue(d, "max_simultaneous_streams", settings.MaxSimultaneousStreams)
	resourcedata.SetNillableValue(d, "max_configurable_screen_recording_streams", settings.MaxConfigurableScreenRecordingStreams)
	resourcedata.SetNillableValue(d, "regional_recording_storage_enabled", settings.RegionalRecordingStorageEnabled)
	resourcedata.SetNillableValue(d, "recording_playback_url_ttl", settings.RecordingPlaybackUrlTtl)
	resourcedata.SetNillableValue(d, "recording_batch_download_url_ttl", settings.RecordingBatchDownloadUrlTtl)
	resourcedata.SetNillableValue(d, "stop_recording_when_only_external_participants", settings.StopRecordingWhenOnlyExternalParticipants)
}
