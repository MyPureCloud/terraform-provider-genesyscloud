resource "genesyscloud_journey_segment" "example_journey_segment_resource" {
  display_name            = "example journey segment name"
  color                   = "#008000"
  should_display_to_agent = false
  context {
    patterns {
      criteria {
        key                = "geolocation.postalCode"
        values             = ["something"]
        operator           = "equal"
        should_ignore_case = true
        entity_type        = "visit"
      }
    }
  }
  journey {
    patterns {
      criteria {
        key                = "page.hostname"
        values             = ["something_else"]
        operator           = "equal"
        should_ignore_case = false
      }
      count        = 1
      stream_type  = "Web"
      session_type = "web"
      event_name   = "EventName"
    }
  }
}