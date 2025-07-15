resource "genesyscloud_task_management_worktype_status_transition" "backlog" {
  worktype_id                     = genesyscloud_task_management_worktype.example_worktype.id
  status_id                       = genesyscloud_task_management_worktype_status.backlog.id
  destination_status_ids          = [genesyscloud_task_management_worktype_status.open.id, genesyscloud_task_management_worktype_status.working.id, genesyscloud_task_management_worktype_status.closed.id]
  default_destination_status_id   = genesyscloud_task_management_worktype_status.open.id
  status_transition_delay_seconds = 86500
  status_transition_time          = "04:20:00"

}
resource "genesyscloud_task_management_worktype_status_transition" "open" {
  worktype_id                     = genesyscloud_task_management_worktype.example_worktype.id
  status_id                       = genesyscloud_task_management_worktype_status.open.id
  destination_status_ids          = [genesyscloud_task_management_worktype_status.working.id, genesyscloud_task_management_worktype_status.waiting.id, genesyscloud_task_management_worktype_status.backlog.id, genesyscloud_task_management_worktype_status.resolved.id, genesyscloud_task_management_worktype_status.closed.id]
  default_destination_status_id   = genesyscloud_task_management_worktype_status.working.id
  status_transition_delay_seconds = 86500
  status_transition_time          = "04:20:00"

}

resource "genesyscloud_task_management_worktype_status_transition" "working" {
  worktype_id                     = genesyscloud_task_management_worktype.example_worktype.id
  status_id                       = genesyscloud_task_management_worktype_status.working.id
  destination_status_ids          = [genesyscloud_task_management_worktype_status.open.id, genesyscloud_task_management_worktype_status.waiting.id, genesyscloud_task_management_worktype_status.resolved.id, genesyscloud_task_management_worktype_status.closed.id]
  default_destination_status_id   = genesyscloud_task_management_worktype_status.waiting.id
  status_transition_delay_seconds = 86500
  status_transition_time          = "04:20:00"
}

resource "genesyscloud_task_management_worktype_status_transition" "waiting" {
  worktype_id                     = genesyscloud_task_management_worktype.example_worktype.id
  status_id                       = genesyscloud_task_management_worktype_status.waiting.id
  destination_status_ids          = [genesyscloud_task_management_worktype_status.working.id, genesyscloud_task_management_worktype_status.resolved.id, genesyscloud_task_management_worktype_status.closed.id]
  default_destination_status_id   = genesyscloud_task_management_worktype_status.working.id
  status_transition_delay_seconds = 86500
  status_transition_time          = "04:20:00"
}

resource "genesyscloud_task_management_worktype_status_transition" "resolved" {
  worktype_id            = genesyscloud_task_management_worktype.example_worktype.id
  status_id              = genesyscloud_task_management_worktype_status.resolved.id
  destination_status_ids = [genesyscloud_task_management_worktype_status.open.id, genesyscloud_task_management_worktype_status.backlog.id]
}

resource "genesyscloud_task_management_worktype_status_transition" "closed" {
  worktype_id            = genesyscloud_task_management_worktype.example_worktype.id
  status_id              = genesyscloud_task_management_worktype_status.closed.id
  destination_status_ids = [genesyscloud_task_management_worktype_status.open.id, genesyscloud_task_management_worktype_status.backlog.id]
}
