resource "genesyscloud_telephony_providers_edges_edge_group" "test_edge_group" {
  name                 = "test edge group"
  description          = "test description"
  managed              = false
  hybrid               = false
  phone_trunk_base_ids = [genesyscloud_telephony_providers_edges_trunkbasesettings.trunk.id]
}