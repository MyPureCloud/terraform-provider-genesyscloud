data "genesyscloud_flow" "default_inqueue_flow" {
  name = "Default In-Queue Flow"
}

data "genesyscloud_flow" "default_voicemail_flow" {
  name = "Default Voicemail Flow"
  type = "voicemail"
}
