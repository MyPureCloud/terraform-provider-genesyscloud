resource "genesyscloud_task_management_worktype_status" "backlog" {
  worktype_id = genesyscloud_task_management_worktype.example_worktype.id
  name        = "Backlog Status"
  description = "Description of Backlog status"
  category    = "Open"
  default     = true
}
resource "genesyscloud_task_management_worktype_status" "open" {
  worktype_id = genesyscloud_task_management_worktype.example_worktype.id
  name        = "Open Status"
  description = "Description of open status"
  category    = "Open"
  default     = false
}

resource "genesyscloud_task_management_worktype_status" "working" {
  worktype_id = genesyscloud_task_management_worktype.example_worktype.id
  name        = "Working Status"
  description = "Description of working status"
  category    = "InProgress"
  default     = false
}

resource "genesyscloud_task_management_worktype_status" "waiting" {
  worktype_id = genesyscloud_task_management_worktype.example_worktype.id
  name        = "Waiting Status"
  description = "Description of wait status"
  category    = "Waiting"
  default     = false
}

resource "genesyscloud_task_management_worktype_status" "resolved" {
  worktype_id = genesyscloud_task_management_worktype.example_worktype.id
  name        = "Resolved Status"
  description = "Description of working status"
  category    = "Closed"
}

resource "genesyscloud_task_management_worktype_status" "closed" {
  worktype_id = genesyscloud_task_management_worktype.example_worktype.id
  name        = "Closed Status"
  description = "Closed statue indicates no longer working, but not resolved"
  category    = "Closed"
}
