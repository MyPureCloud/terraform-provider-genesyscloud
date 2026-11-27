resource "genesyscloud_responsemanagement_library" "example_library" {
  name        = "Example library name"
  division_id = data.genesyscloud_auth_division_home.home.id
}

data "genesyscloud_auth_division_home" "home" {}
