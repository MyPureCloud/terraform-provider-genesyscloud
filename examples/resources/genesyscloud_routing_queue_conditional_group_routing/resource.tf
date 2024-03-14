resource "genesyscloud_routing_queue_conditional_group_routing" "example-name" {
  queue_id = genesyscloud_routing_queue.example-queue.id
  rules {
    operator        = "LessThanOrEqualTo"
    metric          = "EstimatedWaitTime"
    condition_value = 0
    wait_seconds    = 20
    groups {
      member_group_id   = ""
      member_group_type = ""
    }
  }
  rules {
    operator        = "GreaterThanOrEqualTo"
    metric          = "EstimatedWaitTime"
    condition_value = 5
    wait_seconds    = 15
    groups {
      member_group_id   = ""
      member_group_type = ""
    }
  }
}