data "genesyscloud_task_management_worktype_flow_datebased_rule" "datebased_rule_data" {
  worktype_id = genesyscloud_task_management_worktype.example.id
  name        = "DateBased Rule"
}