locals {
  dependencies = {
    resource = [
      "../../data-sources/genesyscloud_auth_division_home/data-source.tf",
      "../genesyscloud_user/resource.tf",
    ]
  }
}
