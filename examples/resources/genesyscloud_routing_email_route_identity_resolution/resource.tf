resource "genesyscloud_routing_email_route_identity_resolution" "example_route_identity_resolution" {
  domain_name        = genesyscloud_routing_email_route.example_route.domain_id
  route_id           = genesyscloud_routing_email_route.example_route.id
  resolve_identities = true
}