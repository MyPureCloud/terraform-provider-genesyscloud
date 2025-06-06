resource "genesyscloud_journey_segment" "terraform_test_-TEST-CASE-" {
  # required
  display_name            = "terraform_test_-TEST-CASE-"
  color                   = "#008000"
  should_display_to_agent = false
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
  # optional
  description = "Test description"
}
