resource "genesyscloud_auth_division" "marketing" {
  name        = "Marketing"
  description = "Custom Division for Marketing"
}
resource "genesyscloud_auth_division" "home" {
  name        = "Home"
  description = "Division for Home"
}
