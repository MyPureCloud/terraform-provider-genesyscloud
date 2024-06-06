data "genesyscloud_journey_action_map" "terraform_test_-TEST-CASE-" {
  name = "terraform_test_-TEST-CASE-_to_find"

  depends_on = [genesyscloud_journey_action_map.terraform_test_-TEST-CASE-]
}

resource "genesyscloud_journey_action_map" "terraform_test_-TEST-CASE-" {
  display_name          = "terraform_test_-TEST-CASE-_to_find"
  trigger_with_segments = [genesyscloud_journey_segment.terraform_test_-TEST-CASE-_action_map_dependency.id]
  activation {
    type = "immediate"
  }
  action {
    media_type = "webMessagingOffer"
  }
  start_date = "2022-07-04T12:00:00.000000"
}

resource "genesyscloud_journey_segment" "terraform_test_-TEST-CASE-_action_map_dependency" {
  display_name            = "terraform_test_-TEST-CASE-_action_map_dependency"
  color                   = "#008000"
  scope                   = "Session"
  should_display_to_agent = true
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
