resource "genesyscloud_journey_segment" "terraform_test_-TEST-CASE-" {
  is_active               = false
  display_name            = "terraform_test_-TEST-CASE-_updated"
  color                   = "#308000"
  scope                   = "Session"
  should_display_to_agent = true
  context {
    patterns {
      criteria {
        key                = "geolocation.region"
        values             = ["something1"]
        operator           = "containsAll"
        should_ignore_case = false
        entity_type        = "visit"
      }
    }
  }
  journey {
    patterns {
      criteria {
        key                = "page.title"
        values             = ["Title"]
        operator           = "notEqual"
        should_ignore_case = true
      }
      count        = 1
      stream_type  = "Web"
      session_type = "web"
    }
  }
}
