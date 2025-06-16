resource "genesyscloud_journey_action_map" "example_journey_action_map" {
  display_name          = "journey_action_map_1"
  start_date            = "2023-01-02T15:04:05.000000"
  trigger_with_segments = [genesyscloud_journey_segment.example_journey_segment_resource.id]
  action {
    media_type = "architectFlow"
    architect_flow_fields {
      architect_flow_id = genesyscloud_flow.inbound_call_flow.id
    }
  }
  activation {
    type = "immediate"
  }
}
