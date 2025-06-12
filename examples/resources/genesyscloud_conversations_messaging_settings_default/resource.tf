resource "genesyscloud_conversations_messaging_settings_default" "example_default_settings" {
  setting_id = genesyscloud_conversations_messaging_settings.example_settings.id
}
