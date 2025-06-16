resource "genesyscloud_idp_salesforce" "salesforce" {
  name         = "Salesforce"
  certificates = [local.salesforce_certificate]
  issuer_uri   = "https://example.com"
  target_uri   = "https://example.com/login"
}
