data "genesyscloud_integration_action" "integrationAction" {
  name = "example integration action name"
}

# Disambiguate a static (built-in) data action by also specifying the parent integration id.
# This is useful when the same static action name (e.g. "Get User") exists across multiple
# integration instances and the exporter has emitted both as data sources.
data "genesyscloud_integration_action" "staticAction" {
  name           = "Get User"
  integration_id = genesyscloud_integration.my_integration.id
}
