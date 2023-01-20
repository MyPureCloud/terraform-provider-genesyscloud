resource "genesyscloud_routing_utilization" "org-utililzation" {
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
}