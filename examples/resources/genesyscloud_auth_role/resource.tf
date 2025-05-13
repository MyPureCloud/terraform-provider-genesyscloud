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

resource "genesyscloud_auth_role" "workitems_role" {
  name        = "Task Management Work Items Role"
  description = "Custom Role for Task Management Work Items "
  permissions = ["workitems"]
  permission_policies {
    domain      = "workitems"
    entity_name = "workbin"
    action_set  = ["view", "add", "edit", "delete"]
  }
  permission_policies {
    domain      = "workitems"
    entity_name = "workitem"
    action_set  = ["view", "add", "edit", "delete"]
  }
}
