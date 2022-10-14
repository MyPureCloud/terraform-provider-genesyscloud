resource "genesyscloud_telephony_providers_edges_extension_pool" "example_extension_pool" {
  start_number = "1000"
  end_number   = "1099"
  description  = "Description of the Extension range"
}
