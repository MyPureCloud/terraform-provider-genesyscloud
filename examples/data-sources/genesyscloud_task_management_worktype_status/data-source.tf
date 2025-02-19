data "genesyscloud_task_management_worktype_status" "status_sample" {
  worktype_id = genesyscloud_task_management_worktype.example.id
  name        = "Worktype status"
}