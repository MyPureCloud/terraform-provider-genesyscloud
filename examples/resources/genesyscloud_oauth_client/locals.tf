locals {
  dependencies = [
    "../genesyscloud_auth_role/resource.tf",
    "../../data-sources/genesyscloud_auth_division_home/data-source.tf",
    "../genesyscloud_user_roles/resource.tf",
  ]
}
