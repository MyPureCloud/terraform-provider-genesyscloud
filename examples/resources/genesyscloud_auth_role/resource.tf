resource "genesyscloud_auth_role" "agent_role" {
  name        = "Agent Role"
  description = "Custom Role for Agents"
  permissions = ["group_creation"]
  permission_policies {
    domain      = "quality"
    entity_name = "evaluation"
    action_set  = ["add", "edit"]
    conditions {
      conjunction = "AND"
      terms {
        variable_name = "Conversation.queues"
        operator      = "EQ"
        operands {
          type     = "QUEUE"
          queue_id = genesyscloud_routing_queue.example_queue.id
        }
      }
    }
  }
}
