locals {
  dependencies = {
    resource = [
      "../genesyscloud_telephony_providers_edges_site/resource.tf",
      "../genesyscloud_telephony_providers_edges_trunkbasesettings/resource.tf",
    ]
  }
}
