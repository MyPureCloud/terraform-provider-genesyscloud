resource "genesyscloud_telephony_providers_edges_phonebasesettings" "example_phonebasesettings" {
  name               = "example phone base settings"
  description        = "Sample description"
  phone_meta_base_id = "inin_webrtc_softphone.json"
  properties = jsonencode({
    "phone_label" = {
      "value" = {
        "instance" = "PureCloud WebRTC Phone"
      }
    },
    "phone_maxLineKeys" = {
      "value" = {
        "instance" = 1
      }
    },
    "phone_media_codecs" = {
      "value" = {
        "instance" = [
          "audio/opus"
        ]
      }
    },
    "phone_media_dscp" = {
      "value" = {
        "instance" = 46
      }
    },
    "phone_sip_dscp" = {
      "value" = {
        "instance" = 24
      }
    }
  })
  capabilities {
    registers             = false
    provisions            = false
    dual_registers        = false
    no_cloud_provisioning = false
    allow_reboot          = false
    hardware_id_type      = "mac"
    no_rebalance          = false
    media_codecs = [
      "audio/opus"
    ]
    cdm = true
  }
}

