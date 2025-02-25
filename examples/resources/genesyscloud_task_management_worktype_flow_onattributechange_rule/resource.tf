resource "genesyscloud_task_management_worktype_flow_onattributechange_rule" "onattributechange_rule_data" {
  worktype_id = genesyscloud_task_management_worktype.example.id
  name        = "OnAttributeChange Rule"
  condition {
    attribute = "statusId"
    new_value = genesyscloud_task_management_worktype_status.example_status_1.id
    old_value = genesyscloud_task_management_worktype_status.example_status_2.id
  }
}
