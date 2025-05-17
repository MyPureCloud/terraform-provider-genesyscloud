resource "genesyscloud_telephony_providers_edges_trunkbasesettings" "example_trunkbasesettings" {
  name               = "example trunk base settings"
  description        = "my example trunk base settings"
  trunk_meta_base_id = "external_sip_pcv_byoc_carrier.json"
  trunk_type         = "EXTERNAL"
  inbound_site_id    = genesyscloud_telephony_providers_edges_site.site.id
  managed            = false
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

