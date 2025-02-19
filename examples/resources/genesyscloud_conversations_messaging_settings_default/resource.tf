resource "genesyscloud_conversations_messaging_settings_default" "example-default-messaging-settings" {
  setting_id = data.genesyscloud_conversations_messaging_settings.example-messaging-settings.id
}