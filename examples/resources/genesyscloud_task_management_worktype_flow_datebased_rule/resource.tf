resource "genesyscloud_task_management_worktype_flow_datebased_rule" "datebased_rule" {
  worktype_id = genesyscloud_task_management_worktype.example_worktype.id
  name        = "DateBased Rule"
  condition {
    attribute                      = "dateDue"
    relative_minutes_to_invocation = -10
  }
}
