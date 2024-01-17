resource "genesyscloud_team" "example_team" {
  name        = "My Team"
  description = "Example Team"
  division_id = genesyscloud_auth_division.example_division.id
  member_ids  = []
}
