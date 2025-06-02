resource "genesyscloud_responsemanagement_response" "example_responsemanagement_response" {
  name        = "Sample response name"
  library_ids = [genesyscloud_responsemanagement_library.example_library.id]
  texts {
    content      = "Sample text content"
    content_type = "text/plain" // Possible values: text/plain, text/html
  }
  interaction_type = "chat" // Possible values: chat, email, twitter
  substitutions {
    id            = "sample_id"
    description   = "Sample description"
    default_value = "Sample default value"
  }
  substitutions_schema_id = jsonencode({
    "type" = "object",
    "required" = [
      "status"
    ],
    "properties" = {
      "status" = {
        "type" = "string"
      }
      "outobj" = {
        "type" = "object",
        "properties" = {
          "objstr" = {
            "type" = "string"
          }
        }
      }
    }
  })
  response_type = "MessagingTemplate" // Possible values: MessagingTemplate, CampaignSmsTemplate, CampaignEmailTemplate, Footer
  messaging_template {
    whats_app {
      name      = "Sample name"
      namespace = "Sample namespace"
      language  = "en_US"
    }
  }
  asset_ids = [genesyscloud_responsemanagement_responseasset.example_asset.id]
}

resource "genesyscloud_responsemanagement_response" "example_responsemanagement_response_footer" {
  library_ids = [genesyscloud_responsemanagement_library.example_library.id]
  name        = "Sample response footer"
  footer {
    type                 = "Signature"
    applicable_resources = ["Campaign"]
  }
  response_type = "Footer"
  texts {
    content      = "<div style=\"font-size: 12pt; font-family: helvetica, arial;\"><p>Sincerely, Foo</p></div>"
    content_type = "text/html"
  }
}

resource "genesyscloud_responsemanagement_response" "example_responsemanagement_response_sms" {
  library_ids = [genesyscloud_responsemanagement_library.example_library.id]
  name        = "Sample response SMS"

  response_type = "CampaignSmsTemplate"
  texts {
    content      = "SMS text messages rates may apply"
    content_type = "text/plain"
  }
}
