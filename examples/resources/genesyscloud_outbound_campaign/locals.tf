locals {
  dependencies = {
    resource = [
      "../genesyscloud_outbound_contact_list/resource.tf",
      "../genesyscloud_telephony_providers_edges_site/resource.tf",
      "../genesyscloud_outbound_callanalysisresponseset/resource.tf",
    ]
  }
}
