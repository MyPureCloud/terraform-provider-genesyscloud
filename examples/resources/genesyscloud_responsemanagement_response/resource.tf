resource "genesyscloud_responsemanagement_response" "example_responsemanagement_response" {
  name        = ""
  library_ids = [genesyscloud_TODO_FILL_IN_RESOURCE_TYPE.library_1.id, genesyscloud_TODO_FILL_IN_RESOURCE_TYPE.library_2.id]
  texts {
    content      = ""
    content_type = "text/plain" // Possible values: text/plain, text/html
  }
  interaction_type = "chat" // Possible values: chat, email, twitter
  substitutions {
    description   = ""
    default_value = ""
  }
  substitutions_schema_id = genesyscloud_TODO_FILL_IN_RESOURCE_TYPE.substitutions_schema.id
  response_type           = "MessagingTemplate" // Possible values: MessagingTemplate, CampaignSmsTemplate, CampaignEmailTemplate
  messaging_template {
    whats_app {
      name      = ""
      namespace = ""
      language  = ""
    }
  }
  asset_ids = [genesyscloud_TODO_FILL_IN_RESOURCE_TYPE.asset_1.id, genesyscloud_TODO_FILL_IN_RESOURCE_TYPE.asset_2.id]
}