resource "genesyscloud_task_management_worktype_status" "worktype_status" {
  worktype_id                     = genesyscloud_task_management_worktype.example.id
  name                            = "Open Status"
  description                     = "Description of open status"
  category                        = "Open"
  destination_status_ids          = [genesyscloud_task_management_worktype_status.status1.id, genesyscloud_task_management_worktype_status.status2.id]
  default_destination_status_id   = genesyscloud_task_management_worktype_status.status1.id
  status_transition_delay_seconds = 86500
  status_transition_time          = "04:20:00"
  default                         = false
}
