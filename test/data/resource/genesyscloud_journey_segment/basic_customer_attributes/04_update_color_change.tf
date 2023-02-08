resource "genesyscloud_journey_segment" "terraform_test_-TEST-CASE-" {
  display_name            = "terraform_test_-TEST-CASE-_updated"
  color                   = "#318234"
  scope                   = "Customer"
  should_display_to_agent = false
  external_segment {
    id     = "111"
    name   = "external segment updated name"
    source = "AdobeExperiencePlatform"
  }
}
