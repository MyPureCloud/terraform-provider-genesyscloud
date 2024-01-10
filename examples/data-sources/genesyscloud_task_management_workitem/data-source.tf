data "genesyscloud_task_management_workitem" "example_workitem" {
  name = "My Workitem"

  // Requires either or both of the following fields
  workbin_id  = genesyscloud_routing_workbin.example.id
  worktype_id = genesyscloud_routing_worktype.example.id
}
