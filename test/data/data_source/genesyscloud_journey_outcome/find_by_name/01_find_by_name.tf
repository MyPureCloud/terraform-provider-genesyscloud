data "genesyscloud_journey_outcome" "terraform_test_-TEST-CASE-" {
  name = "terraform_test_-TEST-CASE-_to_find"

  depends_on = [genesyscloud_journey_outcome.terraform_test_-TEST-CASE-]
}

resource "genesyscloud_journey_outcome" "terraform_test_-TEST-CASE-" {
  display_name = "terraform_test_-TEST-CASE-_to_find"
}
