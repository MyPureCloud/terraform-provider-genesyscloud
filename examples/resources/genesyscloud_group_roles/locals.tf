locals {
  dependencies = {
    resource = [
      "../genesyscloud_group/resource.tf",
      "../genesyscloud_auth_role/resource.tf",
      "../genesyscloud_auth_division/resource.tf"
    ]
  }
}
