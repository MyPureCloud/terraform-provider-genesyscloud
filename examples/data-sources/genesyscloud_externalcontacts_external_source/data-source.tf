data "genesyscloud_externalcontacts_external_source" "external_source" {
  name   = "example-source-123"
  active = true
  link_configuration {
    uri_template = "https://some.host/{{externalId.value}}"
  }
}
