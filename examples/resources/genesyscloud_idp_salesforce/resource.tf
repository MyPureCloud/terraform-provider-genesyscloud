resource "genesyscloud_idp_salesforce" "salesforce" {
  certificates = ["MIIDgjCCAmoCCQCY7/3Fvy+CmDA..."]
  issuer_uri   = "https://example.com"
  target_uri   = "https://example.com/login"
}
