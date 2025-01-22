resource "genesyscloud_task_management_worktype_flow_oncreate_rule" "oncreate_rule" {
  worktype_id = genesyscloud_task_management_worktype.example.id
  name        = "OnCreate Rule"
}
