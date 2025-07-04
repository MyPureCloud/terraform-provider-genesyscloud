resource "genesyscloud_auth_role" "simple_agent_role" {
  name        = "Agent Role"
  description = "Custom Role for Agents"
  permissions = ["group_creation"]
  permission_policies {
    domain      = "directory"
    entity_name = "*"
    action_set  = ["add", "edit"]
  }
}
