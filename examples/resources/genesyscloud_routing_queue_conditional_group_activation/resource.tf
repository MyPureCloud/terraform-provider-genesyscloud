// To enable this resource, set ENABLE_STANDALONE_CGA as an environment variable
// WARNING: This resource will overwrite any conditional group activation rules that already on the queue
// For this reason, all conditional group activation rules for a queue should be managed solely by this resource
resource "genesyscloud_routing_queue_conditional_group_activation" "example_queue_cga" {
  queue_id = genesyscloud_routing_queue.example_queue.id

  pilot_rule {
    condition_expression = "C1"
    conditions {
      simple_metric {
        metric = "EstimatedWaitTime"
      }
      operator = "GreaterThan"
      value    = 30
    }
  }

  rules {
    condition_expression = "C1 or C2"
    conditions {
      simple_metric {
        metric   = "EstimatedWaitTime"
        queue_id = genesyscloud_routing_queue.example_queue.id
      }
      operator = "GreaterThan"
      value    = 60
    }
    conditions {
      simple_metric {
        metric   = "IdleAgentCount"
        queue_id = genesyscloud_routing_queue.example_queue2.id
      }
      operator = "LessThan"
      value    = 2
    }
    groups {
      member_group_id   = genesyscloud_routing_skill_group.example_skill_group.id
      member_group_type = "SKILLGROUP"
    }
  }
}
