resource "genesyscloud_user_roles" "example_user_roles" {
  user_id = genesyscloud_user.example_user.id
  roles {
    role_id      = genesyscloud_auth_role.simple_agent_role.id
    division_ids = [data.genesyscloud_auth_division_home.home.id, genesyscloud_auth_division.marketing.id]
  }
}
