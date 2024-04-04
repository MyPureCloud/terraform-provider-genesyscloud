// WARNING: This resource will overwrite any conditional group routing rules that already on the queue
// For this reason, all conditional group routing rules for a queue should be managed solely by this resource
resource "genesyscloud_routing_queue_conditional_group_routing" "example-name" {
  queue_id  = genesyscloud_routing_queue.example-queue.id
  domain_id = genesyscloud_routing_email_domain.main.id
  route_id  = genesyscloud_routing_email_route.support.id
}