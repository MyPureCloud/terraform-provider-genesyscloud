resource "genesyscloud_group_roles" "group_roles" {
  group_id = genesyscloud_group.example_group.id
  roles {
    role_id      = genesyscloud_auth_role.simple_agent_role.id
    division_ids = [genesyscloud_auth_division.marketing.id]
  }
}
