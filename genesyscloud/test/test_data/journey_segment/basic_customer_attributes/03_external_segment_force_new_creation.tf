resource "genesyscloud_journey_segment" "terraform_test_test_case" {
  display_name            = "terraform_test_journey_segment_updated_2"
  color                   = "#308000"
  scope                   = "Customer"
  should_display_to_agent = false
  external_segment {
    id     = "111"
    name   = "external segment updated 2 name"
    source = "AdobeExperiencePlatform"
  }
}