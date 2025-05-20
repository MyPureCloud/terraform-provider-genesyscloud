resource "genesyscloud_telephony_providers_edges_phone" "example_phone" {
  name                   = "example phone"
  state                  = "active"
  site_id                = genesyscloud_telephony_providers_edges_site.site.id
  phone_base_settings_id = genesyscloud_telephony_providers_edges_phonebasesettings.example_phonebasesettings.id

  line_properties {
    line_address = ["+13175550001"]
  }

  web_rtc_user_id = genesyscloud_user.example_user.id

  depends_on = [
    genesyscloud_telephony_providers_edges_did_pool.example_did_pool
  ]
}
