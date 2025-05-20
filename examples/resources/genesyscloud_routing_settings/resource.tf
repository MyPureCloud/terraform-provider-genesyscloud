resource "genesyscloud_routing_settings" "my_settings" {
  reset_agent_on_presence_change = true
  contactcenter {
    remove_skills_from_blind_transfer = true
  }
  transcription {
    transcription                      = "EnabledQueueFlow"
    transcription_confidence_threshold = 0
    low_latency_transcription_enabled  = true
    content_search_enabled             = true
  }
}
