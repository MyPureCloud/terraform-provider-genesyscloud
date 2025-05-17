locals {
  dependencies = {
    resource = [
      "../genesyscloud_group/resource.tf",
      "../genesyscloud_routing_queue/resource.tf",
    ]
  }
  environment_vars = {
    ENABLE_STANDALONE_CGR = true
  }

}
