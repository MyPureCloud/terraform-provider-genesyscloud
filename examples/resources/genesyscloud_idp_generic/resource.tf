resource "genesyscloud_idp_generic" "generic" {
  name                     = "Generic Provider"
  certificates             = [local.generic_certificate]
  issuer_uri               = "https://example.com"
  target_uri               = "https://example.com/login"
  relying_party_identifier = "unique-id-from-provider"
  logo_image_data          = filebase64("${local.working_dir.idp_generic}/logo.svg")
  endpoint_compression     = false
  name_identifier_format   = "urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified"
}
