resource "genesyscloud_journey_segment" "terraform_test_test_case" {
  display_name            = "terraform_test_journey_segment_updated"
  color                   = "#308000"
  scope                   = "Session"
  should_display_to_agent = true
  journey {
    patterns {
      criteria {
        key                = "attributes.bleki.value"
        values             = ["Blabla"]
        operator           = "notEqual"
        should_ignore_case = true
      }
      count        = 1
      stream_type  = "Web"
      session_type = "web"
      event_name   = "OtherEventName"
    }
  }
}