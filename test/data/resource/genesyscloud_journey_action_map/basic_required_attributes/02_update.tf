resource "genesyscloud_journey_action_map" "terraform_test_-TEST-CASE-" {
  display_name          = "terraform_test_-TEST-CASE-_updated"
  trigger_with_segments = [genesyscloud_journey_segment.terraform_test_-TEST-CASE-_action_map_dependency.id]
  activation {
    type             = "delay"
    delay_in_seconds = 60
  }
  action {
    media_type = "architectFlow"
    architect_flow_fields {
      architect_flow_id = genesyscloud_flow.terraform_test_-TEST-CASE-_action_map_dependency.id
    }
  }
  start_date = "2022-07-05T15:30:00.000000"

  depends_on = [
    genesyscloud_journey_segment.terraform_test_-TEST-CASE-_action_map_dependency,
    genesyscloud_flow.terraform_test_-TEST-CASE-_action_map_dependency
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

resource "genesyscloud_flow" "terraform_test_-TEST-CASE-_action_map_dependency" {
  filepath          = "http://localhost:8112/-TEST-CASE-_journey_action_map_dependency_flow.yaml"
  file_content_hash = "ef6f1d11c4829dfef241f86bbf7238c6612f1448609370c6afbd614e8602c3f9"
  substitutions     = {
    flow_name            = "terraform_test_-TEST-CASE-_flow_name"
    default_language     = "en-us"
    greeting             = "Hello World"
    menu_disconnect_name = "Disconnect"
  }
}
