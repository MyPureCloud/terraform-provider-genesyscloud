resource "genesyscloud_outbound_dnclist" "dnc_list" {
  name            = "Example DNC List"
  dnc_source_type = "rds_custom"
  # login_id        = "1VC392SER23T1534DS23TGFR43JS63D7FS78G88TR9A9"
  dnc_codes = ["B", "F", "S"]
}
