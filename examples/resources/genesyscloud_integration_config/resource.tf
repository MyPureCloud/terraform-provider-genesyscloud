resource "genesyscloud_integration_config" "example_config" {
  integration_id = genesyscloud_integration.example_embedded_client_integration.id
  name           = "Example Integration Config"
  notes          = "Managed by Terraform"
  credentials = {
    basicAuth = genesyscloud_integration_credential.example_basicauth_credential.id
  }
}
