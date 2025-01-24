data "genesyscloud_task_management_worktype_flow_onattributechange_rule" "onattributechange_rule_data" {
  worktype_id = genesyscloud_task_management_worktype.example.id
  name        = "OnAttributeChange Rule"
}