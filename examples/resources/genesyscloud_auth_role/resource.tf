resource "genesyscloud_auth_role" "agent_role" {
  name        = "Agent Role"
  description = "Custom Role for Agents"
  permissions = ["group_creation"]
  permission_policies {
    domain      = "directory"
    entity_name = "user"
    action_set  = ["add", "edit"]
  }
}
