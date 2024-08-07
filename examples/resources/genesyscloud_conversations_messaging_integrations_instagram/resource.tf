resource "genesyscloud_conversations_messaging_integrations_instagram" "test_sample" {
  name                 = "Test Integration Instagram"
  page_access_token    = "1234567890"
  messaging_setting_id = "2c4e3b8e-3c9f-45c9-82cd-4bb54c8f18f0"
  supported_content_id = "019c37a7-ccb4-4966-b1d7-ddb20399f7ab"
}

resource "genesyscloud_conversations_messaging_integrations_instagram" "test_sample" {
  name                 = "Test Integration Instagram"
  user_access_token    = "1234567890"
  page_id              = "1"
  messaging_setting_id = "2c4e3b8e-3c9f-45c9-82cd-4bb54c8f18f0"
  supported_content_id = "019c37a7-ccb4-4966-b1d7-ddb20399f7ab"
}