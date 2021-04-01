resource "genesyscloud_idp_onelogin" "onelogin" {
  certificates = ["MIIDgjCCAmoCCQCY7/3Fvy+CmDA..."]
  issuer_uri   = "https://example.com"
  target_uri   = "https://example.com/login"
}
