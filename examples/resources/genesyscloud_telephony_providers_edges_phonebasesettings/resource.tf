resource "genesyscloud_telephony_providers_edges_phonebasesettings" "examplePhoneBaseSettings" {
  name               = "example phone base settings"
  description        = "Sample description"
  phone_meta_base_id = "generic_sip.json"
  properties = jsonencode({
    "phone_label" = {
      "value" = {
        "instance" = "Generic SIP Phone"
      }
    }
    "phone_maxLineKeys" = {
      "value" = {
        "instance" = 1
      }
    }
    "phone_mwi_enabled" = {
      "value" = {
        "instance" = true
      }
    }
    "phone_mwi_subscribe" = {
      "value" = {
        "instance" = true
      }
    }
    "phone_standalone" = {
      "value" = {
        "instance" = false
      }
    }
    "phone_stations" = {
      "value" = {
        "instance" = ["station 1"]
      }
    }
  })
}

