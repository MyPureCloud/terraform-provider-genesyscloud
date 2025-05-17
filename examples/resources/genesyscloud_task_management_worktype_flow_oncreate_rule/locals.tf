locals {
  dependencies = {
    resource = [
      "../genesyscloud_task_management_worktype/resource.tf"
    ]
  }
  skip_if = {
    products_missing_all = ["workitems"]
  }
}
