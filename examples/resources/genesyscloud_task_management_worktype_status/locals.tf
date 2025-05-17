locals {
  dependencies = {
    resource = [
      "../../data-sources/genesyscloud_auth_division_home/data-source.tf",
      "../genesyscloud_task_management_worktype/resource.tf",
    ]
  }
  skip_if = {
    products_missing_all = ["workitems"]
  }
}
