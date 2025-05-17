locals {
  dependencies = {
    resource = [
      "../genesyscloud_group/resource.tf",
      "../genesyscloud_integration_credential/resource.tf"
    ]
  }
}
