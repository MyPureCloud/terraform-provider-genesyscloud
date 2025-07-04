resource "genesyscloud_processautomation_trigger" "example-trigger" {
  name       = "Example Trigger"
  topic_name = "v2.detail.events.conversation.{id}.customer.end"
  enabled    = true
  target {
    id   = genesyscloud_flow.workflow_flow.id
    type = "Workflow"
    workflow_target_settings {
      data_format = "TopLevelPrimitives"
    }
  }
  match_criteria = jsonencode([
    {
      "jsonPath" : "mediaType",
      "operator" : "Equal",
      "value" : "CHAT"
    }
  ])
  event_ttl_seconds = 60
  description       = "description of trigger"
}
