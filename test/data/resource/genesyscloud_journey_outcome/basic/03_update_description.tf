resource "genesyscloud_journey_outcome" "terraform_test_-TEST-CASE-" {
  is_active    = false
  display_name = "terraform_test_-TEST-CASE-_updated"
  description  = "updating only the description of journey outcome"
  is_positive  = false
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
