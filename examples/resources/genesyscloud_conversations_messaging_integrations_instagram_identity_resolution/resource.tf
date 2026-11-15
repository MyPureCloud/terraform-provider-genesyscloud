resource "genesyscloud_conversations_messaging_integrations_instagram_identity_resolution" "test_sample" {
  instagram_integration_id = genesyscloud_conversations_messaging_integrations_instagram.test_sample.id
  resolve_identities       = true
  division_id              = data.genesyscloud_auth_division_home.home.id
}
