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
  trigger_with_outcome_quantile_conditions {
    outcome_id                  = genesyscloud_journey_outcome.terraform_test_-TEST-CASE-_action_map_dependency.id
    max_quantile_threshold      = 0.666
    fallback_quantile_threshold = 0.125
  }
  # optional
  end_date   = "2022-08-01T10:30:00.999000"
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
