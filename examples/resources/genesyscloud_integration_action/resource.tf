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
