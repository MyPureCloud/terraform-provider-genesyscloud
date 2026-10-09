resource "genesyscloud_conversations_messaging_integrations_whatsapp_identity_resolution" "test_messaging_whatsapp" {
  whatsapp_integration_id = genesyscloud_conversations_messaging_integrations_whatsapp.test_messaging_whatsapp.id
  resolve_identities      = true
  division_id             = data.genesyscloud_auth_division_home.home.id
}
