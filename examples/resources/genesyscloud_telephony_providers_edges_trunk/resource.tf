resource "genesyscloud_telephony_providers_edges_trunk" "test_trunk" {
  trunk_base_settings_id = genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings.id
  edge_group_id          = genesyscloud_telephony_providers_edges_edge_group.edgeGroup.id
}