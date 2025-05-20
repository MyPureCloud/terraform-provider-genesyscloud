resource "genesyscloud_routing_queue" "example_queue" {
  name        = "Example Queue"
  division_id = data.genesyscloud_auth_division_home.home.id
  description = "This is an example description"
  groups      = [genesyscloud_group.example_group.id]
}
