resource "genesyscloud_idp_adfs" "adfs" {
  name                     = "ADFS"
  certificates             = [local.adfs_certificate]
  issuer_uri               = "https://example.com"
  target_uri               = "https://example.com/login"
  relying_party_identifier = "unique-id-from-adfs"
  disabled                 = true
}
