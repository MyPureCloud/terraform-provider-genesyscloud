resource "genesyscloud_user_roles" "user-roles" {
  user_id = genesyscloud_user.user.id
  roles {
    role_id      = genesyscloud_auth_role.custom-role.id
    division_ids = [genesyscloud_auth_division.marketing.id]
  }
}