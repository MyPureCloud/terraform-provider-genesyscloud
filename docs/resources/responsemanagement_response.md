---
page_title: "genesyscloud_responsemanagement_response Resource - terraform-provider-genesyscloud"
subcategory: ""
description: |-
  Genesys Cloud responsemanagement response
---
# genesyscloud_responsemanagement_response (Resource)

<!-- This document is automatically generated. Do not edit manually. Make changes to the schema, examples, or apis.md files in examples/resources/ and run 'make docs' to regenerate. -->

Genesys Cloud responsemanagement response

## API Usage

The following Genesys Cloud APIs are used by this resource. Ensure your OAuth Client has been granted the necessary scopes and permissions to perform these operations:

* [GET /api/v2/responsemanagement/libraries](https://developer.genesys.cloud/devapps/api-explorer#get-api-v2-responsemanagement-libraries)
* [GET /api/v2/responsemanagement/responses](https://developer.genesys.cloud/devapps/api-explorer#get-api-v2-responsemanagement-responses)
* [POST /api/v2/responsemanagement/responses](https://developer.genesys.cloud/devapps/api-explorer#post-api-v2-responsemanagement-responses)
* [DELETE /api/v2/responsemanagement/responses/{responseId}](https://developer.genesys.cloud/devapps/api-explorer#delete-api-v2-responsemanagement-responses--responseId-)
* [GET /api/v2/responsemanagement/responses/{responseId}](https://developer.genesys.cloud/devapps/api-explorer#get-api-v2-responsemanagement-responses--responseId-)
* [PUT /api/v2/responsemanagement/responses/{responseId}](https://developer.genesys.cloud/devapps/api-explorer#put-api-v2-responsemanagement-responses--responseId-)

## Permissions and Scopes

The following permissions are required to use this resource:

* `responses:library:view`
* `responses:response:add`
* `responses:response:delete`
* `responses:response:edit`
* `responses:response:view`

The following OAuth scopes are required to use this resource:

* `response-management`
* `response-management:readonly`


## Example Usage

```terraform
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `library_ids` (Set of String) One or more libraries response is associated with. Changing the library IDs will result in the resource being recreated
- `name` (String) Name of the responsemanagement response

### Optional

- `asset_ids` (Set of String) Assets used in the response
- `footer` (Block Set, Max: 1) Footer template identifies the Footer type and its footerUsage (see [below for nested schema](#nestedblock--footer))
- `form` (Block List, Max: 1) Form template definition for responseType.Form. Requires response_type to be set to 'Form'. (see [below for nested schema](#nestedblock--form))
- `interaction_type` (String) The interaction type for this response.
- `messaging_template` (Block Set, Max: 1) An optional messaging template definition for responseType.MessagingTemplate. (see [below for nested schema](#nestedblock--messaging_template))
- `response_type` (String) The response type represented by the response.
- `substitutions` (Block Set) Details about any text substitutions used in the texts for this response. (see [below for nested schema](#nestedblock--substitutions))
- `substitutions_schema_id` (String) Metadata about the text substitutions in json schema format.
- `texts` (Block Set) One or more texts associated with the response. Required for the Standard, Footer, MessagingTemplate and CampaignEmailTemplate response types. Not used by the Form response type, which ignores texts and always returns an empty list, so setting this alongside 'form' results in a permanent diff. (see [below for nested schema](#nestedblock--texts))

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--footer"></a>
### Nested Schema for `footer`

Optional:

- `applicable_resources` (List of String) Specifies the canned response template where the footer can be used.Valid values: Campaign.
- `type` (String) Specifies the type represented by Footer.Valid values: Signature.


<a id="nestedblock--form"></a>
### Nested Schema for `form`

Required:

- `form_description` (String) Description of the form.
- `form_pages` (Block List, Min: 1, Max: 20) Pages of the form. Must contain between 1 and 20 pages. (see [below for nested schema](#nestedblock--form--form_pages))
- `received_message` (Block List, Min: 1, Max: 1) Message displayed when the response is received. (see [below for nested schema](#nestedblock--form--received_message))
- `reply_message` (Block List, Min: 1, Max: 1) Message displayed as reply. (see [below for nested schema](#nestedblock--form--reply_message))
- `show_summary` (Boolean) Whether to show a summary after form completion.

Optional:

- `introduction` (Block List, Max: 1) Introduction section of the form. (see [below for nested schema](#nestedblock--form--introduction))

<a id="nestedblock--form--form_pages"></a>
### Nested Schema for `form.form_pages`

Required:

- `page_components` (Block List, Min: 1) Components on this page. (see [below for nested schema](#nestedblock--form--form_pages--page_components))
- `subtitle` (String) Subtitle of the page.
- `title` (String) Title of the page.

<a id="nestedblock--form--form_pages--page_components"></a>
### Nested Schema for `form.form_pages.page_components`

Required:

- `form_component_type` (String) Type of the component. The matching component block must be set.

Optional:

- `date_picker` (Block List, Max: 1) Date picker configuration. Required when form_component_type is 'DatePicker'. (see [below for nested schema](#nestedblock--form--form_pages--page_components--date_picker))
- `input` (Block List, Max: 1) Input field configuration. Required when form_component_type is 'Input'. (see [below for nested schema](#nestedblock--form--form_pages--page_components--input))
- `list_picker` (Block List, Max: 1) List picker configuration. Required when form_component_type is 'ListPicker'. (see [below for nested schema](#nestedblock--form--form_pages--page_components--list_picker))
- `wheel_picker` (Block List, Max: 1) Wheel picker configuration. Required when form_component_type is 'WheelPicker'. (see [below for nested schema](#nestedblock--form--form_pages--page_components--wheel_picker))

<a id="nestedblock--form--form_pages--page_components--date_picker"></a>
### Nested Schema for `form.form_pages.page_components.date_picker`

Required:

- `date_display_format` (String) Date display format. For example: 'dayMonthYear', 'monthDayYear' or 'yearMonthDay'.

Optional:

- `subtitle` (String) Subtitle of the date picker.
- `title` (String) Title of the date picker.


<a id="nestedblock--form--form_pages--page_components--input"></a>
### Nested Schema for `form.form_pages.page_components.input`

Required:

- `is_multiple_line` (Boolean) Whether the input supports multiple lines.
- `is_required` (Boolean) Whether the input is required.

Optional:

- `auto_complete_type` (String) A string value representing the keyboard and system information about the expected semantic meaning for the content that users enter.
- `keyboard_type` (String) Type of keyboard to be shown.
- `placeholder_text` (String) Placeholder text for the input.
- `subtitle` (String) Subtitle of the input field.
- `title` (String) Title of the input field.


<a id="nestedblock--form--form_pages--page_components--list_picker"></a>
### Nested Schema for `form.form_pages.page_components.list_picker`

Required:

- `sections` (Block List, Min: 1) Sections in the list picker. At least one section is required. (see [below for nested schema](#nestedblock--form--form_pages--page_components--list_picker--sections))

<a id="nestedblock--form--form_pages--page_components--list_picker--sections"></a>
### Nested Schema for `form.form_pages.page_components.list_picker.sections`

Required:

- `items` (Block List, Min: 2, Max: 100) Items in this section. Must contain between 2 and 100 items. (see [below for nested schema](#nestedblock--form--form_pages--page_components--list_picker--sections--items))
- `multiple_selection` (Boolean) Whether multiple items can be selected.

Optional:

- `title` (String) Title of the section.

<a id="nestedblock--form--form_pages--page_components--list_picker--sections--items"></a>
### Nested Schema for `form.form_pages.page_components.list_picker.sections.items`

Required:

- `title` (String) Title of the item.

Optional:

- `image_url` (String) URL of the image to display.




<a id="nestedblock--form--form_pages--page_components--wheel_picker"></a>
### Nested Schema for `form.form_pages.page_components.wheel_picker`

Required:

- `items` (Block List, Min: 2, Max: 100) Items in the wheel picker. Must contain between 2 and 100 items. (see [below for nested schema](#nestedblock--form--form_pages--page_components--wheel_picker--items))

<a id="nestedblock--form--form_pages--page_components--wheel_picker--items"></a>
### Nested Schema for `form.form_pages.page_components.wheel_picker.items`

Required:

- `title` (String) Title of the item.

Optional:

- `value` (String) Value of the item.





<a id="nestedblock--form--received_message"></a>
### Nested Schema for `form.received_message`

Required:

- `title` (String) Title of the message.

Optional:

- `image_url` (String) URL of the image to display.
- `subtitle` (String) Subtitle of the message.


<a id="nestedblock--form--reply_message"></a>
### Nested Schema for `form.reply_message`

Required:

- `title` (String) Title of the message.

Optional:

- `image_url` (String) URL of the image to display.
- `subtitle` (String) Subtitle of the message.


<a id="nestedblock--form--introduction"></a>
### Nested Schema for `form.introduction`

Required:

- `button_text` (String) Text for the start button.
- `subtitle` (String) Subtitle of the introduction.
- `title` (String) Title of the introduction.

Optional:

- `image_url` (String) URL of the image to display.



<a id="nestedblock--messaging_template"></a>
### Nested Schema for `messaging_template`

Optional:

- `whats_app` (Block Set, Max: 1) Defines a messaging template for a WhatsApp messaging channel (see [below for nested schema](#nestedblock--messaging_template--whats_app))

<a id="nestedblock--messaging_template--whats_app"></a>
### Nested Schema for `messaging_template.whats_app`

Required:

- `language` (String) The messaging template language configured for this template. This is a WhatsApp specific value. For example, 'en_US'
- `name` (String) The messaging template name.

Optional:

- `namespace` (String) The messaging template namespace.



<a id="nestedblock--substitutions"></a>
### Nested Schema for `substitutions`

Required:

- `id` (String) Response substitution identifier.

Optional:

- `default_value` (String) Response substitution default value.
- `description` (String) Response substitution description.


<a id="nestedblock--texts"></a>
### Nested Schema for `texts`

Required:

- `content` (String) Response text content.

Optional:

- `content_type` (String) Response text content type.
- `type` (String) Response text type.

