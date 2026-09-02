---
page_title: "genesyscloud_integration_action Resource - terraform-provider-genesyscloud"
subcategory: ""
description: |-
  Genesys Cloud Integration Actions. See this page for detailed information on configuring Actions: https://help.mypurecloud.com/articles/add-configuration-custom-actions-integrations/
---
# genesyscloud_integration_action (Resource)

<!-- This document is automatically generated. Do not edit manually. Make changes to the schema, examples, or apis.md files in examples/resources/ and run 'make docs' to regenerate. -->

Genesys Cloud Integration Actions. See this page for detailed information on configuring Actions: https://help.mypurecloud.com/articles/add-configuration-custom-actions-integrations/

## API Usage

The following Genesys Cloud APIs are used by this resource. Ensure your OAuth Client has been granted the necessary scopes and permissions to perform these operations:

* [GET /api/v2/integrations](https://developer.genesys.cloud/devapps/api-explorer#get-api-v2-integrations)
* [GET /api/v2/integrations/actions](https://developer.genesys.cloud/devapps/api-explorer#get-api-v2-integrations-actions)
* [POST /api/v2/integrations/actions](https://developer.genesys.cloud/devapps/api-explorer#post-api-v2-integrations-actions)
* [POST /api/v2/integrations/actions/drafts](https://developer.genesys.cloud/devapps/api-explorer#post-api-v2-integrations-actions-drafts)
* [DELETE /api/v2/integrations/actions/{actionId}](https://developer.genesys.cloud/devapps/api-explorer#delete-api-v2-integrations-actions--actionId-)
* [GET /api/v2/integrations/actions/{actionId}](https://developer.genesys.cloud/devapps/api-explorer#get-api-v2-integrations-actions--actionId-)
* [PATCH /api/v2/integrations/actions/{actionId}](https://developer.genesys.cloud/devapps/api-explorer#patch-api-v2-integrations-actions--actionId-)
* [GET /api/v2/integrations/actions/{actionId}/draft](https://developer.genesys.cloud/devapps/api-explorer#get-api-v2-integrations-actions--actionId--draft)
* [GET /api/v2/integrations/actions/{actionId}/draft/function](https://developer.genesys.cloud/devapps/api-explorer#get-api-v2-integrations-actions--actionId--draft-function)
* [PUT /api/v2/integrations/actions/{actionId}/draft/function](https://developer.genesys.cloud/devapps/api-explorer#put-api-v2-integrations-actions--actionId--draft-function)
* [POST /api/v2/integrations/actions/{actionId}/draft/function/upload](https://developer.genesys.cloud/devapps/api-explorer#post-api-v2-integrations-actions--actionId--draft-function-upload)
* [POST /api/v2/integrations/actions/{actionId}/draft/publish](https://developer.genesys.cloud/devapps/api-explorer#post-api-v2-integrations-actions--actionId--draft-publish)
* [GET /api/v2/integrations/actions/{actionId}/function](https://developer.genesys.cloud/devapps/api-explorer#get-api-v2-integrations-actions--actionId--function)
* [GET /api/v2/integrations/actions/{actionId}/templates/{fileName}](https://developer.genesys.cloud/devapps/api-explorer#get-api-v2-integrations-actions--actionId--templates--fileName-)
* [GET /api/v2/integrations/{integrationId}](https://developer.genesys.cloud/devapps/api-explorer#get-api-v2-integrations--integrationId-)

## Permissions and Scopes

The following permissions are required to use this resource:

* `bridge:actions:view`
* `integrations:action:add`
* `integrations:action:delete`
* `integrations:action:edit`
* `integrations:action:view`
* `integrations:actionFunction:edit`
* `integrations:actionFunction:view`
* `integrations:integration:view`

The following OAuth scopes are required to use this resource:

* `integrations`
* `integrations:readonly`
* `upload`

## Export Behavior

### Static Data Actions Are Exported as Data Sources

When exporting integration actions via the `genesyscloud_tf_export` resource, **static (built-in) data actions are emitted as `data` blocks rather than `resource` blocks**. Static data actions are the pre-installed system actions that ship with each Genesys Cloud integration; their IDs are prefixed with `static` (for example, `static_e7b86b86-...`).

These actions are owned and managed by Genesys Cloud and cannot be created, updated, or deleted through the public Integration Actions API. Emitting them as managed resources would therefore produce Terraform configuration that fails on apply. Exporting them as data sources lets other resources (for example, Architect flows) reference them by name while leaving lifecycle management to Genesys Cloud.

#### What this means for you

- Custom integration actions that you (or your team) created continue to be exported as `resource "genesyscloud_integration_action"` blocks.
- Static (built-in) data actions are exported as `data "genesyscloud_integration_action"` blocks that look them up by `name` and `integration_id`.
- References to static data actions from other exported resources are automatically rewritten to use the generated data source (for example, `data.genesyscloud_integration_action.<label>.id`).
- The `integration_id` attribute on the data source is optional, but it is emitted during export to disambiguate static actions whose names may repeat across integration instances.

## Example Usage

```terraform
# Basic integration action example (without function configuration)
# This shows a standard integration action that doesn't require custom function code
resource "genesyscloud_integration_action" "example_action" {
  name                   = "Example Action"
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

# Example with function configuration
# Required when the integration type is "function-data-actions" (or category contains "function data action").
# Genesys Cloud cannot download function zips — keep the zip available locally / in your pipeline.
resource "genesyscloud_integration_action" "example_function_action" {
  name                   = "Example Function Action"
  category               = "Function Data Actions"
  integration_id         = genesyscloud_integration.example_gc_data_integration.id
  secure                 = true
  config_timeout_seconds = 20

  contract_input = jsonencode({
    "type" = "object",
    "required" = [
      "inputData"
    ],
    "properties" = {
      "inputData" = {
        "type" = "string"
      }
    }
  })

  contract_output = jsonencode({
    "type" = "object",
    "required" = [
      "result"
    ],
    "properties" = {
      "result" = {
        "type" = "string"
      }
    }
  })

  config_request {
    request_type         = "POST"
    request_url_template = ""
    request_template     = "$${input.rawRequest}"
  }

  function_config {
    description       = "Custom function for data processing"
    handler           = "index.handler"
    runtime           = "nodejs22.x"
    timeout_seconds   = 15 //ranges from 1 to 15 seconds
    file_path         = "${local.working_dir.integration_action}/function.zip"
    file_content_hash = filesha256("${local.working_dir.integration_action}/function.zip")
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `category` (String) Category of action. Can be up to 256 characters long. Function data actions are detected when the associated integration type is 'function-data-actions', or when the category contains 'function data action' (case-insensitive; underscores and hyphens treated as spaces). Function data actions require function_config to be set.
- `contract_input` (String) JSON Schema that defines the body of the request that the client (edge/architect/postman) is sending to the service, on the /execute path. Changing the contract_input attribute will cause the existing integration_action to be dropped and recreated with a new ID.
- `contract_output` (String) JSON schema that defines the transformed, successful result that will be sent back to the caller. Changing the contract_output attribute will cause the existing integration_action to be dropped and recreated with a new ID.
- `integration_id` (String) The ID of the integration this action is associated with. When the integration type is 'function-data-actions', this action is created as a draft, the zip is uploaded, then published. Changing the integration_id attribute will cause the existing integration_action to be dropped and recreated with a new ID.
- `name` (String) Name of the action. Can be up to 256 characters long

### Optional

- `config_request` (Block List, Max: 1) Configuration of outbound request. (see [below for nested schema](#nestedblock--config_request))
- `config_response` (Block List, Max: 1) Configuration of response processing. (see [below for nested schema](#nestedblock--config_response))
- `config_timeout_seconds` (Number) Optional 1-60 second timeout enforced on the execution or test of this action. This setting is invalid for Custom Authentication Actions.
- `function_config` (Block List, Max: 1) Configuration of the function settings. (see [below for nested schema](#nestedblock--function_config))
- `secure` (Boolean) Indication of whether or not the action is designed to accept sensitive data. Changing the secure attribute will cause the existing integration_action to be dropped and recreated with a new ID. Defaults to `false`.

### Read-Only

- `action_type` (String) The type of the integration action. Computed based on the action ID prefix. Values: `static` (built-in actions shipped by Genesys Cloud) or `custom` (user-created actions).
- `id` (String) The ID of this resource.

<a id="nestedblock--config_request"></a>
### Nested Schema for `config_request`

Required:

- `request_type` (String) HTTP method to use for request (GET | PUT | POST | PATCH | DELETE).

Optional:

- `headers` (Map of String) Map of headers in name, value pairs to include in request.
- `request_template` (String) Velocity template to define request body sent to 3rd party service. Any instances of '${' must be properly escaped as '$${'
- `request_url_template` (String) URL that may include placeholders for requests to 3rd party service.


<a id="nestedblock--config_response"></a>
### Nested Schema for `config_response`

Optional:

- `success_template` (String) Velocity template to build response to return from Action. Any instances of '${' must be properly escaped as '$${'.
- `translation_map` (Map of String) Map 'attribute name' and 'JSON path' pairs used to extract data from REST response.
- `translation_map_defaults` (Map of String) Map 'attribute name' and 'default value' pairs used as fallback values if JSON path extraction fails for specified key.


<a id="nestedblock--function_config"></a>
### Nested Schema for `function_config`

Required:

- `file_path` (String) Local path to the zip file containing the function data action code. Genesys Cloud does not allow downloading function zip files, so exports cannot retrieve the binary. After export, supply the zip at the exported path (or update file_path) before apply. See https://help.genesys.cloud/articles/add-function-configuration/

Optional:

- `description` (String) Description of the function.
- `file_content_hash` (String) Hash value of the function zip file content. Used to detect changes. Typically set with filesha256(file_path). Computed by the provider when omitted.
- `handler` (String) The handler function name.
- `runtime` (String) The runtime environment for the function.
- `timeout_seconds` (Number) Timeout in seconds for the function execution. Range allowed 1 to 15. Default value 15 seconds.
- `zip_id` (String) The ID of the uploaded zip file containing the function code.

