resource "genesyscloud_telephony_providers_edges_edge_group" "example_edge_group" {
  name                 = "example edge group"
  description          = "example description"
  managed              = false
  hybrid               = false
  phone_trunk_base_ids = [genesyscloud_telephony_providers_edges_trunkbasesettings.example_trunkbasesettings.id]
}
