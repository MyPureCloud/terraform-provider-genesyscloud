resource "genesyscloud_recording_settings" "example_recording_settings" {
  max_simultaneous_streams                       = 100
  regional_recording_storage_enabled             = true
  recording_playback_url_ttl                     = 60
  recording_batch_download_url_ttl               = 30
  stop_recording_when_only_external_participants = false
}
