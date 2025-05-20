locals {
  dependencies = {
    resource = [
      "../genesyscloud_group/resource.tf",
      "../genesyscloud_integration_credential/resource.tf"
    ]
    simplest_resource = [
      "../genesyscloud_integration_credential/simplest_resource.tf"
    ]
  }
}
