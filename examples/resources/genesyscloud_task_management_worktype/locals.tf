locals {
  dependencies = {
    resource = [
      "../../data-sources/genesyscloud_auth_division_home/data-source.tf",
      "../genesyscloud_task_management_workbin/resource.tf",
      "../genesyscloud_task_management_workitem_schema/resource.tf",
      "../genesyscloud_routing_language/resource.tf",
      "../genesyscloud_routing_queue/resource.tf",
      "../genesyscloud_routing_skill/resource.tf",
      "../genesyscloud_script/resource.tf",
    ]
  }
  skip_if = {
    products_missing_all = ["workitems"]
  }
}
