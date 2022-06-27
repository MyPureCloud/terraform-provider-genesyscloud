resource "genesyscloud_journey_action_map" "terraform_test_-TEST-CASE-" {
  display_name          = "terraform_test_-TEST-CASE-_to_find"
  trigger_with_segments = ["b04e61dd-a488-4661-87f3-ffc884f788b7"]
}

data "genesyscloud_journey_action_map" "terraform_test_-TEST-CASE-" {
  name       = "terraform_test_-TEST-CASE-_to_find"
  depends_on = [genesyscloud_journey_action_map.terraform_test_-TEST-CASE-]
}
