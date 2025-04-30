resource "genesyscloud_conversations_messaging_integrations_whatsapp" "test_messaging_whatsapp" {
  name                         = "Test Integration Whatsapp"
  embedded_signup_access_token = "test_token"
  messaging_setting_id         = genesyscloud_conversations_messaging_settings.example_settings.id
  supported_content_id         = genesyscloud_conversations_messaging_supportedcontent.example_supported_content.id
  activate_whatsapp {
    phone_number = "+13172222222"
    pin          = "1234"
  }
}
