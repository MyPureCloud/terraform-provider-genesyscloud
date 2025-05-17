locals {
  dependencies = {
    resource = [
      "../genesyscloud_telephony_providers_edges_trunkbasesettings/resource.tf",
      "../genesyscloud_telephony_providers_edges_edge_group/resource.tf",
    ]
  }
}
