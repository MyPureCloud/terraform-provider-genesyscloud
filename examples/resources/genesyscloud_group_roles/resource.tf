resource "genesyscloud_group_roles" "group-roles" {
  group_id = genesyscloud_user.group1.id
  roles {
    role_id      = genesyscloud_auth_role.custom-role.id
    division_ids = [genesyscloud_auth_division.marketing.id]
  }
}