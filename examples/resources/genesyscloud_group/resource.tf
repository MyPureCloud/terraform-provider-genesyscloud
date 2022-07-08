resource "genesyscloud_group" "test_group" {
  name          = "Testing Group"
  description   = "Group for Testers"
  type          = "official"
  visibility    = "public"
  rules_visible = true
  owner_ids     = [genesyscloud_user.test-user.id]
  member_ids    = [genesyscloud_user.test-user.id]
  addresses {
    number = "3174181234"
    type   = "GROUPRING"
  }
}
