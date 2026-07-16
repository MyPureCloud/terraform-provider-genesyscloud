resource "genesyscloud_routing_queue_identity_resolution" "example_queue_identity_resolution" {
  queue_id = genesyscloud_routing_queue.example_queue.id
  call_on_behalf_of_queue {
    resolve_identities = true
    division_id        = data.genesyscloud_auth_division_home.home.id
  }
}
