resource "genesyscloud_idp_gsuite" "gsuite" {
  name                     = "Google Workspace"
  certificates             = [local.gsuite_certificate]
  issuer_uri               = "https://example.com"
  target_uri               = "https://example.com/login"
  relying_party_identifier = "unique-id-from-gsuite"
}
