locals {
  working_dir = {
    architect_ivr = "."
  }
  dependencies = {
    resource = [
      "./flows.tf",
      "../genesyscloud_architect_schedulegroups/resource.tf",
      "../genesyscloud_telephony_providers_edges_did_pool/resource.tf"
    ]
  }
}
