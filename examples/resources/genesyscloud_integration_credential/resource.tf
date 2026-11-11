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

// Write-only credential: the secret is sent to Genesys Cloud but is never stored in the plan or the state file.
// Requires Terraform 1.11 or later. fields_wo and fields cannot be used together.
variable "basic_auth_password" {
  type      = string
  sensitive = true
  ephemeral = true
}

resource "genesyscloud_integration_credential" "example_basicauth_write_only_credential" {
  name                 = "example-basicAuth-write-only-credential"
  credential_type_name = "basicAuth"
  // A JSON object of string values. Each credential type has different required fields.
  fields_wo = jsonencode({
    userName = "someUserName"
    password = var.basic_auth_password
  })
  // Terraform cannot see a change to a write-only value. Increase this number whenever the secret changes
  // so that the new value is sent to Genesys Cloud.
  fields_wo_version = 1
}
