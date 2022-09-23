resource "genesyscloud_orgauthorization_pairing" "example_orgauthorization_pairing" {
  user_ids  = [genesyscloud_user.user-1.id, genesyscloud_user.user-2.id]
  group_ids = [genesyscloud_group.group.id]
}