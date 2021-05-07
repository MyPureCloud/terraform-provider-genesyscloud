resource "genesyscloud_integration" "integration1" {
  name           = "test_integration"
  intended_state = "DISABLED"
  integration_type {
    id          = "embedded-client-app"
    name        = "Client Application"
    description = "Embeds third-party webapps via iframe in the Genesys Cloud UI."
    provider    = "clientapps"
    category    = "Client Apps"
  }
  config {
    name = "Premium Client Application Example"
    properties = jsonencode({
      "displayType" : "standalone",
      "sandbox" : "allow-scripts,allow-same-origin,allow-forms,allow-modals",
      "url" : "https://mypurecloud.github.io/purecloud-premium-app/wizard/index.html"
      # Potential groups and queues filter (Need to look up the key name from integration type schema)
      "groups" : [genesyscloud_group.test_group.id]
    })
    advanced    = jsonencode({})
    notes       = "Test config"
    credentials = jsonencode({})
  }
}