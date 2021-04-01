resource "genesyscloud_idp_generic" "generic" {
  name                     = "Generic Provider"
  certificates             = ["MIIDgjCCAmoCCQCY7/3Fvy+CmDA..."]
  issuer_uri               = "https://example.com"
  target_uri               = "https://example.com/login"
  relying_party_identifier = "unique-id-from-provider"
  logo_image_data          = "PHN2ZyB4bWxucz0iaHR0cDovL3d3dy5..."
  endpoint_compression     = false
  name_identifier_format   = "urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified"
}
