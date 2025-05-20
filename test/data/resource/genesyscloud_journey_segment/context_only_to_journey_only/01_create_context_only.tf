resource "genesyscloud_journey_segment" "terraform_test_-TEST-CASE-" {
  display_name            = "terraform_test_-TEST-CASE-"
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
}
