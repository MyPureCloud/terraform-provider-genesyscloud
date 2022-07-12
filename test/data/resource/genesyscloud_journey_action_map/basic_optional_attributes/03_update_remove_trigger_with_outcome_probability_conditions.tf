resource "genesyscloud_journey_action_map" "terraform_test_-TEST-CASE-" {
  # required
  display_name          = "terraform_test_-TEST-CASE-_updated"
  trigger_with_segments = [genesyscloud_journey_segment.terraform_test_-TEST-CASE-_action_map_dependency.id]
  activation {
    type = "immediate"
  }
  action {
    media_type = "webMessagingOffer"
  }
  start_date = "2022-07-04T12:00:00.000000"
  # optional
  page_url_conditions {
    values   = ["some_other_value", "some_other_value_2"]
    operator = "containsAny"
  }
  ignore_frequency_cap = true
  end_date             = "2022-08-01T10:30:00.999000"

  depends_on = [genesyscloud_journey_segment.terraform_test_-TEST-CASE-_action_map_dependency]
}

resource "genesyscloud_journey_segment" "terraform_test_-TEST-CASE-_action_map_dependency" {
  display_name            = "terraform_test_-TEST-CASE-_action_map_dependency"
  color                   = "#008000"
  scope                   = "Customer"
  should_display_to_agent = false
  external_segment {
    id     = "4654654654"
    name   = "external segment name"
    source = "AdobeExperiencePlatform"
  }
}
