resource "genesyscloud_conversations_messaging_integrations_open" "test_resource_open" {
  name                                                 = "Terraform Messaging Integration Open - 1"
  supported_content_id                                 = genesyscloud_conversations_messaging_supportedcontent.content.id
  messaging_setting_id                                 = genesyscloud_conversations_messaging_settings.settings.id
  outbound_notification_webhook_url                    = "https://mock-server.prv-use1.test-pure.cloud/messaging-service/webhook"
  outbound_notification_webhook_signature_secret_token = "skjdhfsjfhjsdsoquwajkad1234"
}