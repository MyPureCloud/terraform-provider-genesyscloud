resource "genesyscloud_tf_export" "export" {
  directory          = "./terraform"
  resource_types     = ["genesyscloud_user"]
  include_state_file = true
}
