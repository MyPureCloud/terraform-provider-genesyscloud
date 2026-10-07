resource "genesyscloud_conversations_messaging_integrations_open_identity_resolution" "test_resource_open" {
  open_integration_id = genesyscloud_conversations_messaging_integrations_open.test_resource_open.id
  resolve_identities  = true
  division_id         = data.genesyscloud_auth_division_home.home.id
  external_source_id  = genesyscloud_externalcontacts_external_source.external_source.id
}
