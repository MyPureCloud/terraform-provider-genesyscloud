resource "genesyscloud_process_automation_trigger" "test-trigger" {
  name                    = "Test Trigger"
  topic_name              = "v2.detail.events.conversation.{id}.customer.end"
  enabled                 = true
  target_id               = "ae1e0cde-875d-4d13-a498-615e7a9fe956"
  target_type             = "Workflow"
  match_criteria          = jsonencode([
    {
      jsonPath: "mediaType",
      operator: "Equal",
      value: "CHAT"
    }
  ])
  event_ttl_seconds       = 60
}