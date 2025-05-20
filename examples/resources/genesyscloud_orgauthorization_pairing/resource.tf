resource "genesyscloud_orgauthorization_pairing" "example_orgauthorization_pairing" {
  user_ids  = [genesyscloud_user.example_user.id, genesyscloud_user.example_user2.id]
  group_ids = [genesyscloud_group.example_group.id]
}
