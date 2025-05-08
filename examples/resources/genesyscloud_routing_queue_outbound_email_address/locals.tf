locals {
  dependencies = [
    "../genesyscloud_routing_queue/resource.tf",
    "../genesyscloud_routing_email_domain/resource.tf",
    "../genesyscloud_routing_email_route/resource.tf",
  ]
  environment_vars = {
    ENABLE_STANDALONE_EMAIL_ADDRESS = true
  }
}
