locals {
  working_dir = {
    architect_ivr = "."
  }
  dependencies = [
    "./flows.tf",
    "../genesyscloud_architect_schedulegroups/resource.tf",
    "../genesyscloud_telephony_providers_edges_did_pool/resource.tf"
  ]
}
