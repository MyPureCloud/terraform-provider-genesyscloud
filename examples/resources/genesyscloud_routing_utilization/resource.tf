resource "genesyscloud_routing_utilization" "org-utilization" {
  call {
    maximum_capacity = 1
    include_non_acd  = true
  }
  callback {
    maximum_capacity          = 2
    include_non_acd           = false
    interruptible_media_types = ["call", "email"]
  }
  chat {
    maximum_capacity          = 3
    include_non_acd           = false
    interruptible_media_types = ["call"]
  }
  email {
    maximum_capacity          = 2
    include_non_acd           = false
    interruptible_media_types = ["call", "chat"]
  }
  message {
    maximum_capacity          = 4
    include_non_acd           = false
    interruptible_media_types = ["call", "chat"]
  }
  label_utilizations {
    label_id         = genesyscloud_routing_utilization_label.red_label.id
    maximum_capacity = 4
  }
  label_utilizations {
    label_id               = genesyscloud_routing_utilization_label.blue_label.id
    maximum_capacity       = 3
    interrupting_label_ids = [genesyscloud_routing_utilization_label.red_label.id]
  }
}