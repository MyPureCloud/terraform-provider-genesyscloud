resource "genesyscloud_user_roles" "example_workitems_user_roles" {
  user_id = genesyscloud_user.example_user.id
  roles {
    role_id      = genesyscloud_auth_role.workitems_role.id
    division_ids = [data.genesyscloud_auth_division_home.home.id]
  }
}
