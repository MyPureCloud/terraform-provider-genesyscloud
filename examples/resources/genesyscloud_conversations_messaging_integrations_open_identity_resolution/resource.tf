resource "genesyscloud_conversations_messaging_integrations_instagram_identity_resolution" "test_resource_open" {
  instagram_integration_id = genesyscloud_conversations_messaging_integrations_instagram.test_resource_open.id
  resolve_identities       = true
  division_id              = data.genesyscloud_auth_division_home.home.id
}
