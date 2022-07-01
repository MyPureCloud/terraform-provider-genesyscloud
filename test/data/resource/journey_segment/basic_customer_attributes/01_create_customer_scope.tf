resource "genesyscloud_journey_segment" "terraform_test_-TEST-CASE-" {
  display_name            = "terraform_test_-TEST-CASE-"
  color                   = "#008000"
  scope                   = "Customer"
  should_display_to_agent = false
  external_segment {
    id     = "4654654654"
    name   = "external segment name"
    source = "AdobeExperiencePlatform"
  }
}
