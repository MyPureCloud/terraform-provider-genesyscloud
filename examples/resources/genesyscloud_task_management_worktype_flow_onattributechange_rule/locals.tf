locals {
  dependencies = {
    resource = [
      "../genesyscloud_task_management_worktype/resource.tf",
      "../genesyscloud_task_management_worktype_status/resource.tf"
    ]
  }
  skip_if = {
    products_missing_all = ["workitems"]
  }
}
