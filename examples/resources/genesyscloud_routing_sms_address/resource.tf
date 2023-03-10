resource "genesyscloud_routing_sms_address" "example_routing_sms_address" {
  name                 = "Address name"
  street               = "Main street"
  city                 = "Toronto"
  region               = "Ontario"
  postal_code          = "AAAAAA"
  country_code         = "CA"
  auto_correct_address = true
}