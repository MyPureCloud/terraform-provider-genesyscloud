locals {
  dependencies = {
    resource = [
      "../genesyscloud_flow/resource.tf",
      "../genesyscloud_architect_ivr/resource.tf"
    ]
  }
}
