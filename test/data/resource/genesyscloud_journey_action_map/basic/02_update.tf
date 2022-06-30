resource "genesyscloud_journey_action_map" "terraform_test_-TEST-CASE-" {
  display_name          = "terraform_test_-TEST-CASE-_updated"
  trigger_with_segments = ["b04e61dd-a488-4661-87f3-ffc884f788b7"]
  trigger_with_event_conditions {
    key          = "some_key_updated"
    values       = ["something_else"]
    operator     = "notEqual"
    stream_type  = "Web"
    session_type = "web"
  }
  trigger_with_outcome_probability_conditions {
    outcome_id          = "1234567789"
    maximum_probability = 7.2
    probability         = 2.5
  }
}
