data "genesyscloud_auth_division_home" "home" {}

output "home_name" {
  value = data.genesyscloud_auth_division_home.home.name
}