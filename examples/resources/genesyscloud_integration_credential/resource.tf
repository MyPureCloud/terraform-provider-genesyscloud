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

resource "genesyscloud_integration_credential" "example_purecloudoauth_wo_credential" {
  name                 = "example-pureCloudOAuthClient-write-only-credential"
  credential_type_name = "pureCloudOAuthClient"
  # Write-only fields are sent to the API on create and update but are never stored
  # in Terraform state. Mutually exclusive with `fields`. Requires Terraform CLI 1.11+.
  fields_wo = jsonencode({
    clientId     = "ASDDHO292DSO2232DA"
    clientSecret = "XXXXXXXXXXXXXX"
  })
  # Write-only values are invisible to the Terraform planner between plans, so bump
  # fields_wo_version whenever the write-only fields change to plan an update.
  fields_wo_version = 1
}
