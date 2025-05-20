locals {
  dependencies = {
    resource = [
      "../genesyscloud_user/resource.tf",
      "../genesyscloud_auth_division/resource.tf",
      "../genesyscloud_auth_role/simplest_resource.tf",
      "../../data-sources/genesyscloud_auth_division_home/data-source.tf",
    ]
    workitems_role = [
      "../genesyscloud_auth_role/workitems_role.tf",
    ]
  }
}
