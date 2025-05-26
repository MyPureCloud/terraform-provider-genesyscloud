resource "genesyscloud_task_management_worktype_status" "status2" {
  worktype_id = genesyscloud_task_management_worktype.example.id
  name        = "Open Status2"
  description = "Description of open status"
  category    = "Open"
  default     = false
}

resource "genesyscloud_task_management_worktype_status" "status1" {
  worktype_id = genesyscloud_task_management_worktype.example.id
  name        = "Open Status1"
  description = "Description of open status"
  category    = "Open"
  default     = false
}

resource "genesyscloud_task_management_worktype_status_transition" "Insurance_Claim_-_1724158455_Analyze_Claim" {
  worktype_id                     = genesyscloud_task_management_worktype.example.id
  status_id                       = genesyscloud_task_management_worktype_status.status2.id
  destination_status_ids          = [genesyscloud_task_management_worktype_status.status1.id]
  default_destination_status_id   = genesyscloud_task_management_worktype_status.status1.id
  status_transition_delay_seconds = 86500
  status_transition_time          = "04:20:00"
}
