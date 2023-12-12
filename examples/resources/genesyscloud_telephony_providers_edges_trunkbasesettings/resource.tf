resource "genesyscloud_telephony_providers_edges_trunkbasesettings" "trunkBaseSettings" {
  name               = "example trunk base settings"
  description        = "my example trunk base settings"
  trunk_meta_base_id = "phone_connections_webrtc.json"
  trunk_type         = "PHONE"
  managed            = false
  inbound_site_id    = "site_id"
  properties = jsonencode({
    "trunk_label" = {
      "value" = {
        "instance" = "example trunk base settings"
      }
    }
    "trunk_max_dial_timeout" = {
      "value" = {
        "instance" = "1m"
      }
    }
    "trunk_transport_sip_dscp_value" = {
      "value" = {
        "instance" = 25
      }
    }
    "trunk_media_disconnect_on_idle_rtp" = {
      "value" = {
        "instance" = false
      }
    }
    "trunk_media_codec" = {
      "value" = {
        "instance" = ["audio/pcmu"]
      }
    }
  })
}

