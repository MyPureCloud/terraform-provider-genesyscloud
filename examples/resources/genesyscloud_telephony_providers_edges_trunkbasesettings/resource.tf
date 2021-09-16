resource "genesyscloud_telephony_providers_edges_trunkbasesettings" "trunkBaseSettings1234" {
  name               = "example trunk base settings"
  description        = "my example trunk base settings"
  trunk_meta_base_id = "phone_connections_webrtc.json"
  trunk_type         = "PHONE"
  managed            = false
  properties {
    trunk_type                                  = "station"
    trunk_label                                 = "example trunk base settings"
    trunk_enabled                               = true
    trunk_max_dial_timeout                      = "1m"
    trunk_max_call_rate                         = "40/5s"
    trunk_transport_sip_dscp_value              = 24
    trunk_transport_tcp_connect_timeout         = 2
    trunk_transport_tcp_connection_idle_timeout = 86400
    trunk_transport_retryable_reason_codes      = "500-599"
    trunk_transport_retryable_cause_codes       = "1-5,25,27,28,31,34,38,41,42,44,46,62,63,79,91,96,97,99,100,103"
    trunk_media_codec                           = ["audio/opus"]
    trunk_media_dtmf_method                     = "RTP Events"
    trunk_media_dtmf_payload                    = 101
    trunk_media_dscp_value                      = 46
    trunk_media_srtp_cipher_suites              = ["AES_CM_128_HMAC_SHA1_80"]
    trunk_media_disconnect_on_idle_rtp          = true
    trunk_diagnostic_capture_enabled            = false
    trunk_language                              = "en-US"
  }
}

