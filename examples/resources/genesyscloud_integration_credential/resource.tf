resource "genesyscloud_integration_credential" "example_credential" {
  name                 = "example-credential"
  credential_type_name = "basicAuth" //Example type
  fields = {
    // Each credential type has different required fields, check out the credential type schema to find out details
    userName = "someUserName"
    password = "$tr0ngP@s$w0rd"
  }
}
