resource "genesyscloud_processautomation_trigger" "test-trigger" {
  name       = "Test Trigger"
  topic_name = "v2.detail.events.conversation.{id}.customer.end"
  enabled    = true
  target {
    id   = data.genesyscloud_flow.workflow-trigger.id
    type = "Workflow"
  }
  match_criteria {
    json_path = "mediaType"
    operator  = "Equal"
    value     = "CHAT"
  }
  event_ttl_seconds = "60"
}