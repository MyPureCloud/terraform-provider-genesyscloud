resource "genesyscloud_agentic_virtual_agent_version" "example_version" {
  agent_id = genesyscloud_agentic_virtual_agent.example_agent.id

  definition {
    role = "You are a customer support agent for ACME Corp."

    instructions = [
      "Always greet the customer warmly",
      "Be concise and helpful",
      "Escalate billing issues to a live agent"
    ]

    guardrails {
      custom {
        instruction = "Never reveal internal system details"
        enabled     = true
      }
    }

    # A Data Action tool that authenticates the user and produces an "AuthResult" output.
    tools {
      type        = "DataAction"
      name        = "Authenticate"
      description = "Authenticate the user and return an auth token"

      target {
        id   = genesyscloud_integration_action.example_auth_action.id
        name = genesyscloud_integration_action.example_auth_action.name
      }

      input_instructions = ["Call this to authenticate the user before answering account questions"]

      output = "AuthResult"
    }

    tools {
      type        = "KnowledgeBase"
      name        = "FAQ Search"
      description = "Search the FAQ knowledge base for answers"

      target {
        id   = genesyscloud_knowledge_knowledgebase.example_kb.id
        name = genesyscloud_knowledge_knowledgebase.example_kb.name
      }

      input_instructions = ["Use when the user asks a question"]

      # Dependency-only input: links the Knowledge tool to the Authenticate tool's output so the
      # Knowledge tool only runs after the user is authenticated. The value is NOT sent to the
      # knowledge service — it exists only to gate the tool. dependency_only must be true for
      # inputs on Knowledge tools, and fallback_to_user must be false for ToolOutput sources.
      inputs {
        target_name      = "auth_token"
        type             = "AuthResult"
        source           = "ToolOutput"
        required         = true
        fallback_to_user = false
        mapping          = jsonencode(["AuthResult", "token"])
        dependency_only  = true
      }

      # Only invoke the Knowledge tool once the auth token is present.
      input_validation {
        type             = "Structured"
        if_condition     = jsonencode({ group = "And", rules = [{ name = "auth_token", operator = "IsNotNull" }] })
        else_instruction = "Ask the customer to authenticate first."
      }

      output_instructions {
        type = "Python"
        when = "len(result) > 0"
        then = "Summarize the answer for the user"
      }
    }

    types {
      name        = "OrderId"
      type        = "string"
      description = "Customer order ID"
      direction   = "Input"
    }

    # AuthResult is an object with a "token" property so the Knowledge tool's dependency-only
    # input mapping path ["AuthResult", "token"] resolves at publish time.
    types {
      name      = "AuthResult"
      type      = "object"
      direction = "Output"
      properties {
        name = "token"
        type = "string"
      }
    }

    events {
      type    = "UserExit"
      message = "Thank you, goodbye!"
    }

    events {
      type    = "Escalation"
      message = "Transferring you to a live agent."
    }

    events {
      type                                = "Guardrails"
      message                             = "I cannot help with that."
      violation_threshold                 = 3
      violation_threshold_crossed_message = "Session ended due to policy violations."
    }

    settings {
      comfort_statement {
        enabled = true
      }

      dynamic_turn_instructions {
        repetition_check {
          type        = "Agent"
          messages    = 2
          similarity  = "Moderate"
          instruction = "Your last two responses were very similar. Vary your language, tone, and structure substantially."
        }

        repetition_check {
          type        = "User"
          messages    = 3
          similarity  = "VeryStrict"
          instruction = "The user keeps asking the same thing. Acknowledge the repetition and try a different approach or offer to escalate."
        }
      }
    }
  }
}

# MCP integrations are managed outside this provider. This resource is disabled so the repository's
# example acceptance test does not call Genesys Cloud with a placeholder integration ID. Copy the
# definition into an active version and replace the target ID with an existing MCP integration ID.
resource "genesyscloud_agentic_virtual_agent_version" "mcp_tool_example" {
  count = 0

  agent_id = genesyscloud_agentic_virtual_agent.example_agent.id

  definition {
    role         = "You are an order support agent for ACME Corp."
    instructions = ["Use the MCP tool to answer order-status questions"]

    tools {
      type        = "MCP"
      name        = "Get Order Status"
      description = "Look up the current status of an existing order"

      # For MCP tools, target.id is the MCP integration ID and target.name is the exact tool name
      # advertised by the remote MCP server.
      target {
        id   = "00000000-0000-0000-0000-000000000000"
        name = "get_order_status"
      }

      input_instructions = ["Use when the customer asks about an existing order"]

      inputs {
        target_name      = "order_id"
        type             = "OrderId"
        source           = "User"
        required         = true
        fallback_to_user = true
      }

      # A required External value can terminate the session instead of prompting the customer.
      # Its type must be undisclosed, and fallback_to_failure is mutually exclusive with
      # fallback_to_user on the same input.
      inputs {
        target_name         = "tenant_token"
        type                = "TenantToken"
        source              = "External"
        required            = true
        fallback_to_failure = true
      }

      output                    = "OrderStatus"
      context_retention_enabled = true

      input_validation {
        type             = "Python"
        if_condition     = "order_id is not None"
        else_instruction = "Ask the customer for an order ID."
      }

      errors {
        type        = "OrderLookupFailed"
        instruction = "Apologize and offer to try the lookup again later."
      }
    }

    types {
      name        = "OrderId"
      type        = "string"
      description = "Customer order ID"
      direction   = "Input"
    }

    types {
      name        = "TenantToken"
      type        = "string"
      description = "Tenant authentication token supplied by start context"
      direction   = "Input"
      undisclosed = true
    }

    types {
      name        = "OrderStatus"
      type        = "string"
      description = "Status returned by the MCP order lookup tool"
      direction   = "Output"
      enum_values = ["Processing", "Shipped", "Delivered"]
    }

    types {
      name                = "OrderLookupFailed"
      type                = "McpError"
      description         = "The remote MCP order lookup failed"
      direction           = "Output"
      error_codes         = ["ToolError", "RemoteMcpError", "McpTimeout", "RateLimited"]
      default_instruction = "Tell the customer that order lookup is temporarily unavailable."
    }
  }
}
