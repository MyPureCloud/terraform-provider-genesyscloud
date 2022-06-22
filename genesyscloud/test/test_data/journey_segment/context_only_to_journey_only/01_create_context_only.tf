resource "genesyscloud_journey_segment" "terraform_test_test_case" {
  display_name            = "terraform_test_journey_segment"
  color                   = "#008000"
  scope                   = "Session"
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
}