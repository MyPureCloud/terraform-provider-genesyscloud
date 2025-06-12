// To enable this resource, set ENABLE_STANDALONE_EMAIL_ADDRESS as an environment variable
resource "genesyscloud_routing_queue_outbound_email_address" "example_queue_oea" {
  queue_id  = genesyscloud_routing_queue.example_queue.id
  domain_id = genesyscloud_routing_email_domain.example_domain_com.domain_id
  route_id  = genesyscloud_routing_email_route.example_route.id
}
