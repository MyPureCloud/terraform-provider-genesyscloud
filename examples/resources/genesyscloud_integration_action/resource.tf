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
# Note: function_config is only required for function data actions (when category = "Genesys Cloud Data Action")
# For regular integration actions, this section can be omitted
resource "genesyscloud_integration_action" "example_function_action" {
  name                   = "Example Function Action"
  category               = "Genesys Cloud Data Action"
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

  function_config {
    description       = "Custom function for data processing"
    handler           = "index.handler"
    runtime           = "nodejs18.x"
    timeout_seconds   = 30
    file_path         = "/path/to/function.zip"
    file_content_hash = "abc123def456..."
    publish           = true
  }
}
