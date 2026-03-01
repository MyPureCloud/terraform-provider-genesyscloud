locals {
  dependencies = {
    resource = [
      "../genesyscloud_routing_skill_group/resource.tf",
      "../genesyscloud_routing_queue/resource.tf",
    ]
    simplest_resource = [
      "../genesyscloud_routing_skill_group/resource.tf",
      "../genesyscloud_routing_queue/simplest_resource.tf",
    ]
  }
  environment_vars = {
    ENABLE_STANDALONE_CGA = true
  }
}
