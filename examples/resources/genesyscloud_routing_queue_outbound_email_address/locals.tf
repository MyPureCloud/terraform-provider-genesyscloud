locals {
  dependencies = {
    resource = [
      "../genesyscloud_routing_queue/resource.tf",
      "../genesyscloud_routing_email_domain/resource.tf",
      "../genesyscloud_routing_email_route/resource.tf",
    ]
    simplest_resource = [
      "../genesyscloud_routing_queue/simplest_resource.tf",
      "../genesyscloud_routing_email_domain/resource.tf",
      "../genesyscloud_routing_email_route/simplest_resource.tf",
    ]
  }
  environment_vars = {
    ENABLE_STANDALONE_EMAIL_ADDRESS = true
  }
}
