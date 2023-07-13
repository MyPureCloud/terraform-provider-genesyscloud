resource "genesyscloud_group" "sample_group" {
  name          = "Sample Group"
  description   = "Group for Testers"
  type          = "official"
  visibility    = "public"
  rules_visible = true
  owner_ids     = [genesyscloud_user.test-user.id]
  member_ids    = [genesyscloud_user.test-user.id]
  addresses {
    number = "+13174181234"
    type   = "GROUPRING"
  }
}
