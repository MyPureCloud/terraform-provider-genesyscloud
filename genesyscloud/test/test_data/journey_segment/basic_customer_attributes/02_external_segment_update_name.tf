resource "genesyscloud_journey_segment" "terraform_test_test_case" {
  display_name            = "terraform_test_journey_segment_updated"
  color                   = "#308000"
  scope                   = "Customer"
  should_display_to_agent = false
  external_segment {
    id     = "4654654654"
    name   = "external segment updated name"
    source = "AdobeExperiencePlatform"
  }
}