resource "genesyscloud_idp_onelogin" "onelogin" {
  name         = "OneLogin"
  certificates = [local.onelogin_certificate]
  issuer_uri   = "https://example.com"
  target_uri   = "https://example.com/login"
}
