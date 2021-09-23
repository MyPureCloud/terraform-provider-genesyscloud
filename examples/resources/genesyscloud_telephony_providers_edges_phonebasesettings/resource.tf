resource "genesyscloud_telephony_providers_edges_phonebasesettings" "examplePhoneBaseSettings" {
  name               = "example phone base settings"
  description        = "test description"
  phone_meta_base_id = "generic_sip.json"
  properties {
    phone_label         = "Generic SIP Phone"
    phone_max_line_keys = 1
    phone_mwi_enabled   = true
    phone_mwi_subscribe = true
    phone_standalone    = false
    phone_stations      = ["station 1"]
  }
}

