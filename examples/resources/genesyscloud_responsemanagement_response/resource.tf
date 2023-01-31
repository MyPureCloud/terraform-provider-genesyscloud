resource "genesyscloud_responsemanagement_response" "example_responsemanagement_response" {
  name        = "Response name"
  library_ids = [genesyscloud_responsemanagement_library.library_1.id, genesyscloud_responsemanagement_library.library_2.id]
  texts {
    content      = "sample text content"
    content_type = "text/plain" // Possible values: text/plain, text/html
  }
  interaction_type = "chat" // Possible values: chat, email, twitter
  substitutions {
    description   = "Substitution description"
    default_value = "Substitution default value"
  }
  substitutions_schema_id = genesyscloud_TODO_FILL_IN_RESOURCE_TYPE.substitutions_schema.id
  response_type           = "MessagingTemplate" // Possible values: MessagingTemplate, CampaignSmsTemplate, CampaignEmailTemplate
  messaging_template {
    whats_app {
      name      = "Name "
      namespace = "Namespace"
      language  = "en_US"
    }
  }
  asset_ids = [genesyscloud_responsemanagement_responseasset.asset_1.id, genesyscloud_responsemanagement_responseasset.asset_2.id]
}