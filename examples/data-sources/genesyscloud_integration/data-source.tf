data "genesyscloud_integration" "integration" {
  name = "example integration name"
  # Optional: disambiguate when multiple integrations share the same name
  integration_type = "embedded-client-app"
}
