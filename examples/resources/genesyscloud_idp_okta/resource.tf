resource "genesyscloud_idp_okta" "okta" {
  name                = "Okta"
  certificates        = [local.okta_certificate]
  issuer_uri          = "https://example.com"
  target_uri          = "https://example.com/login"
  sign_authn_requests = false
}
