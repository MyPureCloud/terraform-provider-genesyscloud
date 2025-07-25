---
page_title: "genesyscloud_integration_action_draft Resource - terraform-provider-genesyscloud"
subcategory: ""
description: |-
  Genesys Cloud Integration Action Drafts. See this page for detailed information on configuring Actions: https://help.mypurecloud.com/articles/add-configuration-custom-actions-integrations/
---
# genesyscloud_integration_action_draft (Resource)

Genesys Cloud Integration Action Drafts. See this page for detailed information on configuring Actions: https://help.mypurecloud.com/articles/add-configuration-custom-actions-integrations/

## API Usage
The following Genesys Cloud APIs are used by this resource. Ensure your OAuth Client has been granted the necessary scopes and permissions to perform these operations:

* [GET /api/v2/integrations/actions/drafts](https://developer.genesys.cloud/devapps/api-explorer#get-api-v2-integrations-actions-drafts)
* [POST /api/v2/integrations/actions/drafts](https://developer.genesys.cloud/devapps/api-explorer#post-api-v2-integrations-actions-drafts)
* [GET /api/v2/integrations/actions/{actionId}/draft](https://developer.genesys.cloud/devapps/api-explorer#get-api-v2-integrations-actions--actionId--draft)
* [GET /api/v2/integrations/actions/{actionId}/draft/templates/{fileName}](https://developer.genesys.cloud/devapps/api-explorer#get-api-v2-integrations-actions--actionId--draft-templates--fileName-)
* [PATCH /api/v2/integrations/actions/{actionId}/draft](https://developer.genesys.cloud/devapps/api-explorer#patch-api-v2-integrations-actions--actionId--draft)
* [DELETE /api/v2/integrations/actions/{actionId}/draft](https://developer.genesys.cloud/devapps/api-explorer#delete-api-v2-integrations-actions--actionId--draft)

## Example Usage

```terraform
resource "genesyscloud_integration_action_draft" "example-action-draft" {
  name                   = "Example Action Draft"
  category               = "Genesys Cloud Data Action"
  integration_id         = genesyscloud_integration.example_gc_data_integration.id
  secure                 = true
  config_timeout_seconds = 20
  contract_input = jsonencode({
    "type" = "object",
    "required" = [
      "examplestr"
    ],
    "properties" = {
      "examplestr" = {
        "type" = "string"
      },
      "exampleint" = {
        "type" = "integer"
      },
      "examplebool" = {
        "type" = "boolean"
      }
    }
  })
  contract_output = jsonencode({
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
  config_request {
    # Use '$${' to indicate a literal '${' in template strings. Otherwise Terraform will attempt to interpolate the string
    # See https://www.terraform.io/docs/language/expressions/strings.html#escape-sequences
    request_url_template = "https://www.example.com/health/check/services/$${input.service}"
    request_type         = "GET"
    request_template     = "$${input.rawRequest}"
    headers = {
      Cache-Control = "no-cache"
    }
  }
  config_response {
    translation_map = {
      nameValue   = "$.Name"
      buildNumber = "$.Build-Version"
    }
    translation_map_defaults = {
      buildNumber = "UNKNOWN"
    }
    success_template = "{ \"name\": $${nameValue}, \"build\": $${buildNumber} }"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `category` (String) Category of action. Can be up to 256 characters long.
- `contract_input` (String) JSON Schema that defines the body of the request that the client (edge/architect/postman) is sending to the service, on the /execute path. Changing the contract_input attribute will cause the existing integration_action to be dropped and recreated with a new ID.
- `contract_output` (String) JSON schema that defines the transformed, successful result that will be sent back to the caller. Changing the contract_output attribute will cause the existing integration_action to be dropped and recreated with a new ID.
- `integration_id` (String) The ID of the integration this action is associated with. Changing the integration_id attribute will cause the existing integration_action to be dropped and recreated with a new ID.
- `name` (String) Name of the action. Can be up to 256 characters long

### Optional

- `config_request` (Block List, Max: 1) Configuration of outbound request. (see [below for nested schema](#nestedblock--config_request))
- `config_response` (Block List, Max: 1) Configuration of response processing. (see [below for nested schema](#nestedblock--config_response))
- `config_timeout_seconds` (Number) Optional 1-60 second timeout enforced on the execution or test of this action. This setting is invalid for Custom Authentication Actions.
- `secure` (Boolean) Indication of whether or not the action is designed to accept sensitive data. Changing the secure attribute will cause the existing integration_action to be dropped and recreated with a new ID. Defaults to `false`.

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--config_request"></a>
### Nested Schema for `config_request`

Required:

- `request_type` (String) HTTP method to use for request (GET | PUT | POST | PATCH | DELETE).
- `request_url_template` (String) URL that may include placeholders for requests to 3rd party service.

Optional:

- `headers` (Map of String) Map of headers in name, value pairs to include in request.
- `request_template` (String) Velocity template to define request body sent to 3rd party service. Any instances of '${' must be properly escaped as '$${'


<a id="nestedblock--config_response"></a>
### Nested Schema for `config_response`

Optional:

- `success_template` (String) Velocity template to build response to return from Action. Any instances of '${' must be properly escaped as '$${'.
- `translation_map` (Map of String) Map 'attribute name' and 'JSON path' pairs used to extract data from REST response.
- `translation_map_defaults` (Map of String) Map 'attribute name' and 'default value' pairs used as fallback values if JSON path extraction fails for specified key.

