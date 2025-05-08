// To enable this resource, set ENABLE_STANDALONE_CGR as an environment variable
// WARNING: This resource will overwrite any conditional group routing rules that already on the queue
// For this reason, all conditional group routing rules for a queue should be managed solely by this resource
resource "genesyscloud_routing_queue_conditional_group_routing" "example_queue_cgr" {
  queue_id = genesyscloud_routing_queue.example_queue.id
  rules {
    operator        = "LessThanOrEqualTo"
    metric          = "EstimatedWaitTime"
    condition_value = 0
    wait_seconds    = 20
    groups {
      member_group_id   = genesyscloud_group.example_group.id
      member_group_type = "GROUP"
    }
  }
  rules {
    evaluated_queue_id = genesyscloud_routing_queue.example_queue2.id
    operator           = "GreaterThanOrEqualTo"
    metric             = "EstimatedWaitTime"
    condition_value    = 5
    wait_seconds       = 15
    groups {
      member_group_id   = genesyscloud_group.example_group2.id
      member_group_type = "GROUP"
    }
  }
}
