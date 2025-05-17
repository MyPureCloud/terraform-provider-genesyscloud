resource "genesyscloud_telephony_providers_edges_did_pool" "example_did_pool" {
  start_phone_number = "+13175550000"
  end_phone_number   = "+13175560000"
  description        = "Description of the DID range"
  comments           = "Additional comments"
  pool_provider      = "PURE_CLOUD"
}
