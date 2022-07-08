resource "genesyscloud_oauth_client" "example-client" {
  name                          = "Example OAuth Client"
  description                   = "For example purposes only"
  access_token_validity_seconds = 600
  registered_redirect_uris      = ["https://example.com/auth"]
  authorized_grant_type         = "CODE"
  scopes                        = ["users"]
  state                         = "active"
  roles {
    // Roles are only applicable to CLIENT_CREDENTIAL grants
    role_id     = genesyscloud_auth_role.employee.id
    division_id = genesyscloud_auth_division.testing.id
  }
}
