resource "genesyscloud_journey_action_map" "terraform_test_-TEST-CASE-" {
  display_name          = "terraform_test_-TEST-CASE-_updated"
  trigger_with_segments = [genesyscloud_journey_segment.terraform_test_-TEST-CASE-_action_map_dependency.id]
  activation {
    type = "immediate"
  }
  action {
    media_type = "architectFlow"
    architect_flow_fields {
      architect_flow_id = genesyscloud_flow.terraform_test_-TEST-CASE-_action_map_dependency.id
      flow_request_mappings {
        name           = "Name_1"
        attribute_type = "String"
        mapping_type   = "Lookup"
        value          = "session.id"
      }
      flow_request_mappings {
        name           = "Name_2"
        attribute_type = "Integer"
        mapping_type   = "HardCoded"
        value          = "999"
      }
    }
  }
  start_date = "2022-07-04T12:00:00.000000"

  depends_on = [
    genesyscloud_journey_segment.terraform_test_-TEST-CASE-_action_map_dependency,
    genesyscloud_flow.terraform_test_-TEST-CASE-_action_map_dependency
  ]
}

resource "genesyscloud_journey_segment" "terraform_test_-TEST-CASE-_action_map_dependency" {
  display_name            = "terraform_test_-TEST-CASE-_action_map_dependency"
  color                   = "#008000"
  scope                   = "Session"
  should_display_to_agent = true
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

resource "genesyscloud_flow" "terraform_test_-TEST-CASE-_action_map_dependency" {
  filepath          = "http://localhost:8111/-TEST-CASE-_journey_action_map_dependency_flow.yaml"
  file_content_hash = "4e3548b6accc632ca393249769685f029e0af3cb937d10d3ef04993637edddfd"
  substitutions     = {
    flow_name            = "terraform_test_-TEST-CASE-_flow_name"
    default_language     = "en-us"
    greeting             = "Hello World"
    menu_disconnect_name = "Disconnect"
  }
}
