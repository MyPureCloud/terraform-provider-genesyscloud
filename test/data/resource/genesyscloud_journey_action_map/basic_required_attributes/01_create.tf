resource "genesyscloud_journey_action_map" "terraform_test_-TEST-CASE-" {
  display_name          = "terraform_test_-TEST-CASE-"
  trigger_with_segments = ["b04e61dd-a488-4661-87f3-ffc884f788b7"]
  activation {
    type = "immediate"
  }
}
