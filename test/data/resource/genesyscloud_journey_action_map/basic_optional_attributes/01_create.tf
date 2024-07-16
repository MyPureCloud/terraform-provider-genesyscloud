resource "genesyscloud_journey_action_map" "terraform_test_-TEST-CASE-" {
  # required
  display_name          = "terraform_test_-TEST-CASE-"
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
    outcome_id = genesyscloud_journey_outcome.terraform_test_-TEST-CASE-_action_map_dependency.id
    maximum_probability = 0.333
  }
  # optional
  trigger_with_outcome_quantile_conditions {
    outcome_id = genesyscloud_journey_outcome.terraform_test_-TEST-CASE-_action_map_dependency.id
    max_quantile_threshold = 0.333
  }
  page_url_conditions {
    values   = ["some_value"]
    operator = "containsAll"
  }
  ignore_frequency_cap = false
  end_date             = "2022-07-20T19:00:00.000000"

  depends_on = [
    genesyscloud_journey_segment.terraform_test_-TEST-CASE-_action_map_dependency,
    genesyscloud_journey_outcome.terraform_test_-TEST-CASE-_action_map_dependency
  ]
}

resource "genesyscloud_journey_segment" "terraform_test_-TEST-CASE-_action_map_dependency" {
  display_name            = "terraform_test_-TEST-CASE-_action_map_dependency"
  color                   = "#008000"
  scope                   = "Session"
  should_display_to_agent = false
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
