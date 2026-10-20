resource "genesyscloud_conversations_messaging_integrations_apple_identity_resolution" "example" {
  apple_integration_id = genesyscloud_conversations_messaging_integrations_apple.example.id
  resolve_identities   = true
  division_id          = data.genesyscloud_auth_division_home.home.id
}
