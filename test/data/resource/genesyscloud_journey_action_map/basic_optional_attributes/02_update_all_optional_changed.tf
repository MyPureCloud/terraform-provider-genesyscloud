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
  trigger_with_outcome_probability_conditions {
    outcome_id          = genesyscloud_journey_outcome.terraform_test_-TEST-CASE-_action_map_dependency.id
    maximum_probability = 0.666
    probability         = 0.125
  }
  page_url_conditions {
    values   = ["some_other_value", "some_other_value_2"]
    operator = "containsAny"
  }
  ignore_frequency_cap = true
  end_date             = "2022-08-01T10:30:00.999000"

  depends_on = [
    genesyscloud_journey_segment.terraform_test_-TEST-CASE-_action_map_dependency,
    genesyscloud_journey_outcome.terraform_test_-TEST-CASE-_action_map_dependency
  ]
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

resource "genesyscloud_journey_outcome" "terraform_test_-TEST-CASE-_action_map_dependency" {
  is_active    = true
  display_name = "terraform_test_-TEST-CASE-_action_map_dependency"
  description  = "test description of journey outcome"
  is_positive  = true
  journey {
    patterns {
      criteria {
        key                = "page.title"
        values             = ["Title"]
        operator           = "notEqual"
        should_ignore_case = true
      }
      count        = 1
      stream_type  = "Web"
      session_type = "web"
    }
  }
}
