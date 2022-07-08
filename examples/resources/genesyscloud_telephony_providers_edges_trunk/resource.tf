resource "genesyscloud_telephony_providers_edges_trunk" "example_trunk" {
  trunk_base_settings_id = genesyscloud_telephony_providers_edges_trunkbasesettings.trunk-base-settings.id
  edge_group_id          = genesyscloud_telephony_providers_edges_edge_group.edge-group.id
}