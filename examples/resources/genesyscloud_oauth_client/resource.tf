resource "genesyscloud_oauth_client" "example_client_credential" {
  name                          = "Example OAuth Client"
  description                   = "For example purposes only"
  access_token_validity_seconds = 600
  registered_redirect_uris      = ["https://example.com/auth"]
  authorized_grant_type         = "CODE"
  scopes                        = ["users"]
  state                         = "active"
  # roles {
  #   // Roles are only applicable to CLIENT-CREDENTIALS grants
  #   role_id     = genesyscloud_auth_role.agent_role.id
  #   division_id = data.genesyscloud_auth_division_home.home.id
  # }
}
