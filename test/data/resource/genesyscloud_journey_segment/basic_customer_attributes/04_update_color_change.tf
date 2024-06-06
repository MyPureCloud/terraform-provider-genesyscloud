resource "genesyscloud_journey_segment" "terraform_test_-TEST-CASE-" {
  display_name            = "terraform_test_-TEST-CASE-_updated"
  color                   = "#318234"
  should_display_to_agent = false
    scope = "Session"

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
