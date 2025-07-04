locals {
  dependencies = {
    resource = [
      "../../data-sources/genesyscloud_auth_division_home/data-source.tf",
      "../../common/time.tf",
      "../genesyscloud_task_management_worktype/resource.tf",
      "../genesyscloud_task_management_worktype_status/resource.tf",
      "../genesyscloud_routing_language/resource.tf",
      "../genesyscloud_task_management_workbin/resource.tf",
      "../genesyscloud_user/resource.tf",
      "../genesyscloud_user_roles/workitems_role.tf",
      "../genesyscloud_externalcontacts_contact/resource.tf",
      "../genesyscloud_routing_skill/resource.tf",
    ]
  }
  skip_if = {
    products_missing_all = ["workitems"]
  }
}
