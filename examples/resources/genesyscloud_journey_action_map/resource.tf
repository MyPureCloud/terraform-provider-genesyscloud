resource "genesyscloud_journey_action_map" "example_journey_action_map" {
  display_name          = "journey_action_map_1"
  start_date            = "2023-01-02T15:04:05.000000"
  trigger_with_segments = [genesyscloud_journey_segment.segment.id]
  action {
    media_type = "architectFlow"
    architect_flow_fields {
      architect_flow_id = genesyscloud_flow.flow.id
    }
  }
  activation {
    type = "immediate"
  }
}