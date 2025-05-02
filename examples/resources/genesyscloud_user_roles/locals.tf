locals {
  dependencies = [
    "../genesyscloud_user/resource.tf",
    "../genesyscloud_auth_role/resource.tf",
    "../genesyscloud_auth_division/resource.tf",
    "../../data-sources/genesyscloud_auth_division_home/data-source.tf",
  ]
}
