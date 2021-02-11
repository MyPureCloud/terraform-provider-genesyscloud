resource "genesyscloud_user" "test_user" {
  email       = "test@example.com"
  name        = "Test User"
  password    = "initial-password"
  division_id = "505e1036-6f04-405c-a630-de94a8ad2eb8"
}