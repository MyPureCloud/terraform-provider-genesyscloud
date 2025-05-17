locals {
  dependencies = {
    resource = [
      "../genesyscloud_outbound_contact_list/resource.tf",
      "../genesyscloud_routing_queue/resource.tf",
      "../genesyscloud_integration_action/resource.tf",
      "../genesyscloud_routing_wrapupcode/resource.tf",
    ]
  }
}
