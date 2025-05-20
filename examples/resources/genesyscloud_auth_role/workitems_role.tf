resource "genesyscloud_auth_role" "workitems_role" {
  name        = "Task Management Work Items Role"
  description = "Custom Role for Task Management Work Items "
  permissions = ["workitems"]
  permission_policies {
    domain      = "workitems"
    entity_name = "workbin"
    action_set  = ["view", "add", "edit", "delete"]
  }
  permission_policies {
    domain      = "workitems"
    entity_name = "workitem"
    action_set  = ["view", "add", "edit", "delete"]
  }
}
