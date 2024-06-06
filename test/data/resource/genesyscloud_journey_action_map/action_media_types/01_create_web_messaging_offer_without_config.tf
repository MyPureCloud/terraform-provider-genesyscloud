resource "genesyscloud_journey_action_map" "terraform_test_-TEST-CASE-" {
  display_name          = "terraform_test_-TEST-CASE-"
  trigger_with_segments = [genesyscloud_journey_segment.terraform_test_-TEST-CASE-_action_map_dependency.id]
  activation {
    type = "immediate"
  }
  action {
    media_type = "webMessagingOffer"
  }
  start_date = "2022-07-04T12:00:00.000000"

  depends_on = [genesyscloud_journey_segment.terraform_test_-TEST-CASE-_action_map_dependency]
}

resource "genesyscloud_journey_segment" "terraform_test_-TEST-CASE-_action_map_dependency" {
  display_name            = "terraform_test_-TEST-CASE-_action_map_dependency"
  color                   = "#008000"
  scope                   = "Session"
  should_display_to_agent = false
  journey {
      patterns {
        criteria {
          key                = "page.hostname"
          values             = ["something_else"]
          operator           = "equal"
          should_ignore_case = false
        }
        count        = 1
        stream_type  = "Web"
        session_type = "web"
        event_name   = "EventName"
      }
    }
}