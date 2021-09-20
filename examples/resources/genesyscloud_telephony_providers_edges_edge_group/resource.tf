resource "genesyscloud_telephony_providers_edges_edge_group" "test_edge_group" {
  name                 = "test edge group"
  description          = "test description"
  managed              = "false"
  hybrid               = "false"
  phone_trunk_base_ids = ["38ef9294-76c3-4dc8-99f6-3946d9ba6e5d", "a47aa6af-e111-47ac-91d8-3ebbf5d00579"]
}