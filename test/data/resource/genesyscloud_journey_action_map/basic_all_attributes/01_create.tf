resource "genesyscloud_journey_action_map" "terraform_test_-TEST-CASE-" {
  display_name          = "terraform_test_-TEST-CASE-"
  trigger_with_segments = ["b04e61dd-a488-4661-87f3-ffc884f788b7"]
  trigger_with_event_conditions {
    key          = "some_key"
    values       = ["something"]
    operator     = "equal"
    stream_type  = "Web"
    session_type = "web"
  }
  trigger_with_outcome_probability_conditions {
    outcome_id          = "987654321"
    maximum_probability = 7.9
    probability         = 2.1
  }
  page_url_conditions {
    values   = ["url_part_1", "url_part_2"]
    operator = "containsAll"
  }
}
