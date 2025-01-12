resource "genesyscloud_externalcontacts_external_source" "external_source" {
  name   = "some-external-source"
  active = true
  link_configuration {
    uri_template = "https://some.host/{{externalId.value}}"
  }
}
