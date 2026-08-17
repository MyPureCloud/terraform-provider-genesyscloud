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

    tools {
      type        = "KnowledgeBase"
      name        = "FAQ Search"
      description = "Search the FAQ knowledge base for answers"

      target {
        id   = genesyscloud_knowledge_knowledgebase.example_kb.id
        name = genesyscloud_knowledge_knowledgebase.example_kb.name
      }

      input_instructions = ["Use when the user asks a question"]

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
    }
  }
}
