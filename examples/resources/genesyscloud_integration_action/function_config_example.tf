# Example demonstrating the function_config schema for integration actions
# This shows how to configure a custom function with file upload
# 
# IMPORTANT: function_config is only required for function data actions 
# (when category = "function data action")
# For regular integration actions, this section can be omitted

resource "genesyscloud_integration_action" "function_example" {
  name           = "Custom Data Processing Function"
  category       = "function data action"
  integration_id = genesyscloud_integration.example_gc_data_integration.id
  secure         = true

  # Contract defines the input/output schema for the function
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
      },
      "options" = {
        "type" = "object",
        "properties" = {
          "algorithm" = {
            "type"    = "string",
            "default" = "AES-256"
          },
          "timeout" = {
            "type"    = "integer",
            "default" = 30
          }
        }
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
      },
      "metadata" = {
        "type" = "object",
        "properties" = {
          "processingTime" = {
            "type"        = "number",
            "description" = "Time taken to process in milliseconds"
          },
          "algorithm" = {
            "type"        = "string",
            "description" = "Algorithm used for processing"
          }
        }
      }
    }
  })

  # Function configuration block - this is where file_path and file_content_hash are defined
  function_config {
    description       = "Custom data processing function with encryption/decryption capabilities"
    handler           = "src/index.handler"
    runtime           = "nodejs18.x"
    timeout_seconds   = 45
    file_path         = "/path/to/your/function.zip"
    file_content_hash = "sha256:abcdef1234567890..."
    publish           = true # Automatically publish after creation
  }
}

# Example with minimal function configuration
resource "genesyscloud_integration_action" "minimal_function" {
  name           = "Simple Function"
  category       = "Genesys Cloud Data Action"
  integration_id = genesyscloud_integration.example_gc_data_integration.id

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

  function_config {
    handler           = "handler.js"
    runtime           = "python3.9"
    file_path         = "./function.zip"
    file_content_hash = "sha256:1234567890abcdef..."
    # publish defaults to true if not specified
  }
}

# Example showing how to use variables for file paths
resource "genesyscloud_integration_action" "variable_function" {
  name           = "Variable Function Example"
  category       = "Genesys Cloud Data Action"
  integration_id = genesyscloud_integration.example_gc_data_integration.id

  contract_input = jsonencode({
    "type" = "object",
    "properties" = {
      "data" = {
        "type" = "string"
      }
    }
  })

  contract_output = jsonencode({
    "type" = "object",
    "properties" = {
      "processed" = {
        "type" = "string"
      }
    }
  })

  function_config {
    description       = "Function using variable references"
    handler           = "main.handler"
    runtime           = "nodejs16.x"
    timeout_seconds   = 20
    file_path         = var.function_zip_path
    file_content_hash = var.function_zip_hash
    publish           = var.auto_publish
  }
}

# Variables for the variable_function example
variable "function_zip_path" {
  description = "Path to the function zip file"
  type        = string
  default     = "./functions/processor.zip"
}

variable "function_zip_hash" {
  description = "Hash of the function zip file content"
  type        = string
  default     = "sha256:default_hash_here"
}

variable "auto_publish" {
  description = "Whether to automatically publish the action"
  type        = bool
  default     = false
} 