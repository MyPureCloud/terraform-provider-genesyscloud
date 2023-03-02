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
  scope                   = "Customer"
  should_display_to_agent = false
  external_segment {
    id     = "4654654654"
    name   = "external segment name"
    source = "AdobeExperiencePlatform"
  }
}

resource "genesyscloud_journey_action_map" "terraform_test_-TEST-CASE-2" {
  display_name          = "terraform_test_-TEST-CASE-2"
  trigger_with_event_conditions {
  key = "page.title"
  values = ["mytitle"]
  operator = "equal"
  event_name = "page_viewed"
  session_type = "web"
  stream_type = "Web"
  }
  action {
    media_type = "webMessagingOffer"
    is_pacing_enabled = false
    web_messaging_offer_fields {
      offer_text = "Hey how're you?"
    }
  }
  activation {
    type = "immediate"
  }
  start_date = "2023-01-02T15:04:05.000000"
  weight = 2
  is_active = false
  ignore_frequency_cap = false
}