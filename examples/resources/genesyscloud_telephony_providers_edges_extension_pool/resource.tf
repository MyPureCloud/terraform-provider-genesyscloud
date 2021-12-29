resource "genesyscloud_telephony_providers_edges_extension_pool" "test_extension_pool" {
  start_phone_number = "1000"
  end_phone_number   = "1099"
  description        = "Description of the Extension range"
}
