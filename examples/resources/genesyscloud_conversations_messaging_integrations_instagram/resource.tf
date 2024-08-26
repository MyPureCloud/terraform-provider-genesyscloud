resource "genesyscloud_conversations_messaging_integrations_instagram" "test_sample" {
  name                 = "Test Integration Instagram"
  user_access_token    = "1234567890"
  page_id              = "1"
  messaging_setting_id = genesyscloud_conversations_messaging_settings.settings.id
  supported_content_id = genesyscloud_conversations_messaging_supportedcontent.content.id
}