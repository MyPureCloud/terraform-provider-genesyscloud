# Example demonstrating the function_config schema for integration actions
#
# Function data actions are detected when:
#   - the associated integration has integration_type = "function-data-actions", OR
#   - category contains "function data action" (case-insensitive)
#
# Genesys Cloud does not allow downloading function ZIP files. Keep zips in source
# control / your pipeline and point file_path at them. After a CX as Code export,
# copy the zip to the exported path (under function_zips/) or update file_path.

resource "genesyscloud_integration" "example_function_integration" {
  integration_type = "function-data-actions"
  intended_state   = "ENABLED"
  config {
    name       = "Example Function Integration"
    properties = jsonencode({})
    advanced   = jsonencode({})
  }
}

resource "genesyscloud_integration_action" "function_example" {
  name           = "Custom Data Processing Function"
  category       = "Function Data Actions"
  integration_id = genesyscloud_integration.example_function_integration.id
  secure         = true

  contract_input = jsonencode({
    "type" = "object",
    "required" = [
      "inputData",
      "operation"
    ],
    "properties" = {
      "inputData" = {
        "type"        = "string",
        "description" = "The data to be processed"
      },
      "operation" = {
        "type"        = "string",
        "enum"        = ["encrypt", "decrypt", "transform"],
        "description" = "The operation to perform"
      }
    }
  })

  contract_output = jsonencode({
    "type" = "object",
    "required" = [
      "success",
      "result"
    ],
    "properties" = {
      "success" = {
        "type"        = "boolean",
        "description" = "Whether the operation was successful"
      },
      "result" = {
        "type"        = "string",
        "description" = "The processed result"
      }
    }
  })

  config_request {
    request_type         = "POST"
    request_url_template = ""
    request_template     = "$${input.rawRequest}"
  }

  function_config {
    description       = "Custom data processing function"
    handler           = "dist/index.handler"
    runtime           = "nodejs22.x"
    timeout_seconds   = 15
    file_path         = "${path.module}/function.zip"
    file_content_hash = filesha256("${path.module}/function.zip")
  }
}

# Minimal function configuration (hash is optional; provider computes it on apply)
resource "genesyscloud_integration_action" "minimal_function" {
  name           = "Simple Function"
  category       = "Function Data Actions"
  integration_id = genesyscloud_integration.example_function_integration.id

  contract_input = jsonencode({
    "type" = "object",
    "properties" = {
      "message" = {
        "type" = "string"
      }
    }
  })

  contract_output = jsonencode({
    "type" = "object",
    "properties" = {
      "response" = {
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
    handler   = "index.handler"
    runtime   = "nodejs22.x"
    file_path = var.function_zip_path
  }
}

variable "function_zip_path" {
  description = "Path to the function zip file"
  type        = string
  default     = "./functions/processor.zip"
}
