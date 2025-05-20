resource "genesyscloud_integration" "example_gc_data_integration" {
  intended_state   = "ENABLED"
  integration_type = "purecloud-data-actions"
  config {
    credentials = {
      pureCloudOAuthClient = genesyscloud_integration_credential.example_purecloudoauth_credential.id
    }
  }
}
