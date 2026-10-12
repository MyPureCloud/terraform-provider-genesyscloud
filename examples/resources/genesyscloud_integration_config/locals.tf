locals {
  dependencies = {
    resource = [
      "../genesyscloud_integration/resource.tf",
      "../genesyscloud_integration_credential/resource.tf",
    ]
  }
}
