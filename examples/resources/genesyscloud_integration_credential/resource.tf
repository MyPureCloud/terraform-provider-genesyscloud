resource "genesyscloud_integration_credential" "example_basicauth_credential" {
  name                 = "example-basicAuth-credential"
  credential_type_name = "basicAuth"
  fields = {
    // Each credential type has different required fields, check out the credential type schema to find out details
    userName = "someUserName"
    password = "$tr0ngP@s$w0rd"
  }
}
resource "genesyscloud_integration_credential" "example_userDefinedOAuth_credential" {
  name                 = "example-userDefinedOAuth-credential"
  credential_type_name = "userDefinedOAuth"
  fields = {
    // User defined credentials allow any arbitrary key/value pairs
    clientId     = "someId"
    clientSecret = "XXXXXXXXXX"
    loginUrl     = "https://login.example.com"
  }
}

resource "genesyscloud_integration_credential" "example_purecloudoauth_credential" {
  name                 = "example-pureCloudOAuthClient-credential"
  credential_type_name = "pureCloudOAuthClient"
  fields = {
    // Each credential type has different required fields, check out the credential type schema to find out details
    clientId     = "ASDDHO292DSO2232DA"
    clientSecret = "XXXXXXXXXXXXXX"
  }
}
