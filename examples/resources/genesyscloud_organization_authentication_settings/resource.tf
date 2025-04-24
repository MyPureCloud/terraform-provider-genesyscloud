resource "genesyscloud_organization_authentication_settings" "example-authentication-settings" {
  multifactor_authentication_required = true
  domain_allowlist_enabled            = true
  domain_allowlist                    = ["example.com", "example2.com"]
  ip_address_allowlist                = ["0.0.0.0/32"]
  password_requirements {
    minimumLength     = 8
    minimumDigits     = 5
    minimumLetters    = 2
    minimumUpper      = 1
    minimumLower      = 1
    minimumSpecials   = 1
    minimumAgeSeconds = 2
    expirationDays    = 90
  }
}
