resource "genesyscloud_integration" "example_embedded_client_integration" {
  intended_state   = "DISABLED"
  integration_type = "embedded-client-app"
  config {
    name = "example_integration name"
    properties = jsonencode({
      "displayType" = "standalone",
      "sandbox"     = "allow-scripts,allow-same-origin,allow-forms,allow-modals",
      "url"         = "https://mypurecloud.github.io/purecloud-premium-app/wizard/index.html"
      # Potential groups and queues filter (Need to look up the key name from integration type schema)
      "groups" = [genesyscloud_group.example_group.id]
    })
    advanced = jsonencode({})
    notes    = "Test config notes"
  }
}

resource "genesyscloud_integration" "example_rest_integration" {
  intended_state   = "ENABLED"
  integration_type = "custom-rest-actions"
  config {
    credentials = {
      basicAuth = genesyscloud_integration_credential.example_userDefinedOAuth_credential.id
    }
  }
}

resource "genesyscloud_integration" "example_imap_integration" {
  intended_state   = "DISABLED"
  integration_type = "imap-server"
  config {
    credentials = {
      basicAuth = genesyscloud_integration_credential.example_basicauth_credential.id
    }
    name = "example imap integration name"
    properties = jsonencode({
      "imapHost" = "mail.example.com"
      "imapPort" = 993
    })
    advanced = jsonencode({})
    notes    = "Test config notes"
  }
}

resource "genesyscloud_integration" "example_gc_data_integration" {
  intended_state   = "ENABLED"
  integration_type = "purecloud-data-actions"
  config {
    credentials = {
      pureCloudOAuthClient = genesyscloud_integration_credential.example_purecloudoauth_credential.id
    }
  }
}
