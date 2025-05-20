locals {
  dependencies = {
    resource = [
      "../genesyscloud_group/resource.tf",
      "../genesyscloud_auth_role/simplest_resource.tf",
      "../genesyscloud_auth_division/resource.tf"
    ]
  }
}
