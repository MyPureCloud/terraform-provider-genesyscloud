resource "genesyscloud_architect_ivr" "test_ivr" {
  name        = "Sample IVR"
  description = "A sample IVR configuration"
  dnis        = ["+13175550000", "+13175550001"]
}
