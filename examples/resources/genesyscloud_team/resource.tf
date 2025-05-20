resource "genesyscloud_team" "example_team" {
  name        = "My Team"
  description = "Example Team"
  division_id = data.genesyscloud_auth_division_home.home.id
  member_ids = [
    genesyscloud_user.example_user.id
  ]
}
