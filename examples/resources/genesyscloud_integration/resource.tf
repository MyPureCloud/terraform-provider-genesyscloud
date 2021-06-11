resource "genesyscloud_integration" "integration1" {
  intended_state   = "DISABLED"
  integration_type = "embedded-client-app"
  config {
    name = "test_integration name"
    properties = jsonencode({
      "displayType" = "standalone",
      "sandbox"     = "allow-scripts,allow-same-origin,allow-forms,allow-modals",
      "url"         = "https://mypurecloud.github.io/purecloud-premium-app/wizard/index.html"
      # Potential groups and queues filter (Need to look up the key name from integration type schema)
      "groups" = [genesyscloud_group.test_group.id]
    })
    advanced = jsonencode({})
    notes    = "Test config notes"
    credentials = {
      basic_Auth = genesyscloud_credentials.test_cred.id
    }
  }
}