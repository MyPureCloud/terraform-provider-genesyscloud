// To enable this resource, set ENABLE_STANDALONE_EMAIL_ADDRESS as an environment variable
resource "genesyscloud_routing_queue_outbound_email_address" "example-name" {
  queue_id  = genesyscloud_routing_queue.example-queue.id
  domain_id = genesyscloud_routing_email_domain.main.id
  route_id  = genesyscloud_routing_email_route.support.id
}
