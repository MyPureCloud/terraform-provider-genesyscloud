resource "genesyscloud_telephony_providers_edges_did_pool" "test_did_pool" {
  start_phone_number = "+13175550000"
  end_phone_number   = "+13175550000"
  description        = "Description of the DID range"
  comments           = "Additional comments"
  provider           = "PURE_CLOUD"
}
