resource "genesyscloud_telephony_providers_edges_phone" "example_phone" {
  name                   = "example phone"
  state                  = "active"
  site_id                = genesyscloud_telephony_providers_edges_site.site.id
  phone_base_settings_id = genesyscloud_telephony_providers_edges_phonebasesettings.phone-base-settings.id
  line_base_settings_id  = data.genesyscloud_telephony_providers_edges_linebasesettings.line-base-settings.id
  line_addresses         = ["+13175550000"]
  web_rtc_user_id        = genesyscloud_user.user.id

  capabilities {
    provisions            = false
    registers             = false
    dual_registers        = false
    allow_reboot          = false
    no_rebalance          = false
    no_cloud_provisioning = false
    cdm                   = true
    hardware_id_type      = "mac"
    media_codecs          = ["audio/opus"]
  }
}