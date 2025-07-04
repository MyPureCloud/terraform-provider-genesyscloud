resource "genesyscloud_idp_ping" "ping" {
  name                     = "Ping"
  certificates             = [local.ping_certificate]
  issuer_uri               = "https://example.com"
  target_uri               = "https://example.com/login"
  relying_party_identifier = "unique-id-from-ping"
}
