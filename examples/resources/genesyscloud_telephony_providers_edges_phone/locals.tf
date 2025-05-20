locals {
  dependencies = {
    resource = [
      "../genesyscloud_telephony_providers_edges_site/resource.tf",
      "../genesyscloud_telephony_providers_edges_phonebasesettings/resource.tf",
      "../genesyscloud_telephony_providers_edges_did_pool/resource.tf",
      "../genesyscloud_user/resource.tf",
    ]
  }
}
