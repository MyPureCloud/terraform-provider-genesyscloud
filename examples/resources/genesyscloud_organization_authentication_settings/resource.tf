resource "genesyscloud_organization_authentication_settings" "example_authentication_settings" {
  multifactor_authentication_required = false
  domain_allowlist_enabled            = true
  domain_allowlist                    = ["example.com", "genesys.com"]
  ip_address_allowlist                = []
  password_requirements {
    minimum_length      = 8
    minimum_digits      = 1
    minimum_letters     = 0
    minimum_upper       = 1
    minimum_lower       = 1
    minimum_specials    = 1
    minimum_age_seconds = 0
    expiration_days     = 0
  }
}
