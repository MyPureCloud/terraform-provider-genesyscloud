resource "genesyscloud_journey_segment" "terraform_test_-TEST-CASE-" {
  display_name            = "terraform_test_-TEST-CASE-_updated"
  color                   = "#308000"
  scope                   = "Session"
  should_display_to_agent = true
  journey {
    patterns {
      criteria {
        key                = "page.title"
        values             = ["Title"]
        operator           = "notEqual"
        should_ignore_case = true
      }
      criteria {
        key                = "page.keywords"
        values             = ["office", "hubhub"]
        operator           = "containsAny"
        should_ignore_case = true
      }
      count        = 1
      stream_type  = "Web"
      session_type = "web"
    }
    patterns {
      criteria {
        key                = "searchQuery"
        values             = ["Query string"]
        operator           = "notContainsAll"
        should_ignore_case = true
      }
      count        = 1
      stream_type  = "Web"
      session_type = "web"
    }
  }
}
