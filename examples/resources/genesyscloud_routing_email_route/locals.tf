locals {
  dependencies = {
    resource = [
      "../genesyscloud_routing_email_domain/resource.tf",
      "../genesyscloud_routing_queue/resource.tf",
      "../genesyscloud_routing_skill/resource.tf",
      "../genesyscloud_routing_language/resource.tf",
      "../genesyscloud_flow/resource.tf",
      "../genesyscloud_responsemanagement_response/resource.tf",
    ]
    simplest_resource = [
      "../genesyscloud_routing_email_domain/resource.tf"
    ]
  }
}
