resource "genesyscloud_journey_segment" "terraform_test_-TEST-CASE-" {
  display_name            = "terraform_test_-TEST-CASE-"
  color                   = "#008000"
  scope                   = "Session"
  should_display_to_agent = false
  context {
    patterns {
      criteria {
        key                = "geolocation.postalCode"
        values             = ["something"]
        #operator          = "equal"
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
        #operator          = "equal"
        should_ignore_case = false
      }
      count        = 1
      stream_type  = "Web"
      session_type = "web"
    }
  }
}
