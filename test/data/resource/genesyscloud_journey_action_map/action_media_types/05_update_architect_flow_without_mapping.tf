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
  scope                   = "Customer"
  should_display_to_agent = false
  external_segment {
    id     = "4654654654"
    name   = "external segment name"
    source = "AdobeExperiencePlatform"
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
