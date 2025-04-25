resource "genesyscloud_group" "example_group" {
  name          = "Example Group"
  description   = "Group for Testers"
  type          = "official"
  visibility    = "public"
  rules_visible = true
  addresses {
    number = "+13174181234"
    type   = "GROUPRING"
  }
  owner_ids     = [genesyscloud_user.example_user.id]
  member_ids    = [genesyscloud_user.example_user.id]
  roles_enabled = true
  calls_enabled = false
}
