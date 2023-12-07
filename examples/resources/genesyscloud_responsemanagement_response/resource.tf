resource "genesyscloud_responsemanagement_response" "example_responsemanagement_response" {
  name        = "Sample response name"
  library_ids = [genesyscloud_responsemanagement_library.library_1.id, genesyscloud_responsemanagement_library.library_2.id]
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
  footer {
    type                = "Signature"
    applicableResources = ["Campaign"]
  }
  asset_ids = [genesyscloud_responsemanagement_responseasset.asset_1.id, genesyscloud_responsemanagement_responseasset.asset_2.id]
}