resource "genesyscloud_idp_ping" "ping" {
  certificates             = ["MIIDgjCCAmoCCQCY7/3Fvy+CmDA..."]
  issuer_uri               = "https://example.com"
  target_uri               = "https://example.com/login"
  relying_party_identifier = "unique-id-from-ping"
}
