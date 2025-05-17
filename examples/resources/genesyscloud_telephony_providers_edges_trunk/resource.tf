resource "genesyscloud_telephony_providers_edges_trunk" "example_trunk" {
  trunk_base_settings_id = genesyscloud_telephony_providers_edges_trunkbasesettings.example_trunkbasesettings.id
  edge_group_id          = genesyscloud_telephony_providers_edges_edge_group.example_edge_group.id
}
