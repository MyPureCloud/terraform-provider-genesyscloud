locals {
  dependencies = {
    resource = [
      "../genesyscloud_routing_queue/resource.tf",
    ]
    simplest_resource = [
      "../genesyscloud_routing_queue/simplest_resource.tf",
    ]
  }
}
