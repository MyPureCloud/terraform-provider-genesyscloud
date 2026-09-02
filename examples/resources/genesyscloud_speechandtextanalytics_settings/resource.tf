resource "genesyscloud_speechandtextanalytics_settings" "settings" {
  expected_dialects      = ["en-US"]
  text_analytics_enabled = true
  agent_empathy_enabled  = false
}
