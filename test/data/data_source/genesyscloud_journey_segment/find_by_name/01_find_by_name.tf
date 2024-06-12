data "genesyscloud_journey_segment" "terraform_test_-TEST-CASE-" {
  name = "terraform_test_-TEST-CASE-_to_find"

  depends_on = [genesyscloud_journey_segment.terraform_test_-TEST-CASE-]
}

resource "genesyscloud_journey_segment" "terraform_test_-TEST-CASE-" {
  display_name            = "terraform_test_-TEST-CASE-_to_find"
  color                   = "#008000"
  scope                   = "Session"
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
}
