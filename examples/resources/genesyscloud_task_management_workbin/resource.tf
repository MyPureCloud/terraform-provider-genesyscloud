resource "genesyscloud_task_management_workbin" "example_workbin" {
  name        = "My Workbin"
  description = "Example workbin"
  division_id = genesyscloud_auth_division.example_division.id
}
