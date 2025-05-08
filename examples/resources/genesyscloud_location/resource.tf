resource "genesyscloud_location" "hq" {
  name  = "Indy"
  notes = "Main Indy Office"
  address {
    street1  = "7601 Interactive Way"
    city     = "Indianapolis"
    state    = "IN"
    country  = "US"
    zip_code = "46278"
  }
  emergency_number {
    number = "+13173124657"
    type   = "default"
  }
}
