resource "genesyscloud_journey_action_map" "terraform_test_-TEST-CASE-" {
  display_name = "terraform_test_-TEST-CASE-_updated"
  trigger_with_event_conditions {
    key          = "page.title"
    values       = ["mytitle_updated"]
    operator     = "equal"
    event_name   = "page_viewed"
    session_type = "web"
    stream_type  = "Web"
  }
  action {
    media_type        = "webMessagingOffer"
    is_pacing_enabled = false
    web_messaging_offer_fields {
      offer_text = "Hey how're you keeping?"
    }
  }
  activation {
    type = "immediate"
  }
  start_date           = "2023-02-02T15:04:05.000000"
  weight               = 2
  is_active            = false
  ignore_frequency_cap = false
}