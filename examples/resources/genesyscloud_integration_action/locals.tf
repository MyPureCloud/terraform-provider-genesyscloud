locals {
  working_dir = {
    integration_action = "."
  }
  dependencies = {
    resource = [
      "../genesyscloud_integration/resource.tf"
    ]
    simplest_resource = [
      "../genesyscloud_integration/simplest_resource.tf"
    ]
  }
}
