locals {
  dependencies = {
    resource = [
      "../genesyscloud_group/resource.tf",
      "../genesyscloud_routing_queue/resource.tf",
    ]
    simplest_resource = [
      "../genesyscloud_group/resource.tf",
      "../genesyscloud_routing_queue/simplest_resource.tf",
    ]
  }
  environment_vars = {
    ENABLE_STANDALONE_CGR = true
  }

}
