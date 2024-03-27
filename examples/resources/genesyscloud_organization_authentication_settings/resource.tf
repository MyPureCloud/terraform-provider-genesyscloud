resource "genesyscloud_organization_authentication_settings" "example-authentication-settings" {
  multifactorAuthenticationRequired = true
  domainAllowlistEnabled            = true
  domainAllowlist                   = ["example.com", "example2.com"]
  ipAddressAllowlist                = ["0.0.0.0/32"]
  passwordRequirements {
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