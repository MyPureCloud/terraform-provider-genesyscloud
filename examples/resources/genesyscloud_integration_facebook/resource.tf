resource "genesyscloud_integration_facebook" "test_sample1" {
  name                 = "ARBI Org Facebook Integration1"
  page_access_token    = "1234567890"
  messaging_setting_id = genesyscloud_conversations_messaging_settings.example_settings.id
  supported_content_id = genesyscloud_conversations_messaging_supportedcontent.example_supported_content.id
}

resource "genesyscloud_integration_facebook" "test_sample2" {
  name                 = "ARBI Org Facebook Integration2"
  user_access_token    = "1234567890"
  page_id              = "1"
  messaging_setting_id = genesyscloud_conversations_messaging_settings.example_settings.id
  supported_content_id = genesyscloud_conversations_messaging_supportedcontent.example_supported_content.id
}
