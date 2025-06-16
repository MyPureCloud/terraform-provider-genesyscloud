resource "genesyscloud_task_management_workbin" "example_workbin" {
  name        = "My Workbin"
  description = "Example workbin"
  division_id = data.genesyscloud_auth_division_home.home.id
}
