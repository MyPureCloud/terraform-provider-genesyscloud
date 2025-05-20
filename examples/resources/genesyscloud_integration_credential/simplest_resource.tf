
resource "genesyscloud_integration_credential" "example_purecloudoauth_credential" {
  name                 = "example-pureCloudOAuthClient-credential"
  credential_type_name = "pureCloudOAuthClient"
  fields = {
    // Each credential type has different required fields, check out the credential type schema to find out details
    clientId     = "ASDDHO292DSO2232DA"
    clientSecret = "XXXXXXXXXXXXXX"
  }
}
