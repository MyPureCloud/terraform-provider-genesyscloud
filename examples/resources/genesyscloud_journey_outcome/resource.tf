resource "genesyscloud_journey_outcome" "example_journey_outcome" {
  is_active    = true
  display_name = "example journey outcome name"
  description  = "description of journey outcome"
  is_positive  = true
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
}
