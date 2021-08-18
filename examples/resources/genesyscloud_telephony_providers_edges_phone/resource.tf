resource "genesyscloud_telephony_providers_edges_phone" "test_phone" {
  name                   = "test phone"
  state                  = "active"
  site_id                = "48382f1b-fbbe-4232-8dd6-42a4fa70c1b6"
  phone_base_settings_id = "9bae71b4-7ba8-46f1-bb35-710c0c1b225b"
  line_base_settings_id  = "e9069894-b078-4905-b14f-488a6309b82b"
  line_addresses         = ["+13175550000"]
  web_rtc_user_id        = genesyscloud_user.id

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