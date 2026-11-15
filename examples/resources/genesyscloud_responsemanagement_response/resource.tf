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
  response_type = "MessagingTemplate" // Possible values: MessagingTemplate, CampaignSmsTemplate, CampaignEmailTemplate, Footer, Form
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

resource "genesyscloud_responsemanagement_response" "example_responsemanagement_response_form" {
  library_ids = [genesyscloud_responsemanagement_library.example_library.id]
  name        = "Sample response form"

  response_type = "Form"

  // Note: Form responses do not use `texts`. The API ignores it and always returns an
  // empty list, so setting it here would produce a permanent diff.
  form {
    form_description = "A form for customer feedback"
    show_summary     = true

    received_message {
      title    = "Thank you for your feedback"
      subtitle = "We appreciate your input"
    }

    reply_message {
      title    = "Your response has been received"
      subtitle = "We will review your feedback"
    }

    introduction {
      title       = "Customer Feedback Form"
      subtitle    = "Please help us improve our service"
      button_text = "Start Survey"
    }

    // A list picker page: the customer selects from a list of options
    form_pages {
      title    = "Service Rating"
      subtitle = "How would you rate our service?"
      page_components {
        form_component_type = "ListPicker"
        list_picker {
          sections {
            title              = "Select your rating"
            multiple_selection = false
            items {
              title = "Excellent"
            }
            items {
              title = "Good"
            }
            items {
              title = "Poor"
            }
          }
        }
      }
    }

    // A date picker page
    form_pages {
      title    = "Visit Date"
      subtitle = "When did you visit us?"
      page_components {
        form_component_type = "DatePicker"
        date_picker {
          title               = "Select Date"
          subtitle            = "Choose the date of your visit"
          date_display_format = "dayMonthYear" // Possible values: dayMonthYear, monthDayYear, yearMonthDay
        }
      }
    }

    // A free text input page
    form_pages {
      title    = "Additional Comments"
      subtitle = "Please share any additional feedback"
      page_components {
        form_component_type = "Input"
        input {
          title              = "Comments"
          subtitle           = "Your feedback helps us improve"
          placeholder_text   = "Enter your comments here..."
          is_multiple_line   = true
          is_required        = false
          keyboard_type      = "Default" // Possible values: Default, NumberPunctuation, Number, Phone, Email, Decimal, Websearch, URL
          auto_complete_type = "Name"
        }
      }
    }

    // A wheel picker page
    form_pages {
      title    = "Recommendation"
      subtitle = "Would you recommend us to others?"
      page_components {
        form_component_type = "WheelPicker"
        wheel_picker {
          items {
            title = "Definitely"
            value = "definitely"
          }
          items {
            title = "Maybe"
            value = "maybe"
          }
          items {
            title = "Probably Not"
            value = "probably_not"
          }
        }
      }
    }
  }
}
