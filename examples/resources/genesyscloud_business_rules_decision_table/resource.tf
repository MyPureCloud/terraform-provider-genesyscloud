resource "genesyscloud_business_rules_decision_table" "example_decision_table" {
  name        = "Example Decision Table"
  description = "Example Decision Table created by terraform"
  division_id = data.genesyscloud_auth_division_home.home.id
  schema_id   = genesyscloud_business_rules_schema.example_business_rules_schema.id

  columns {
    inputs {
      defaults_to {
        value = "anything"
      }
      expression {
        contractual {
          schema_property_key = "custom_attribute_string"
        }
        comparator = "Equals"
      }
    }

    inputs {
      defaults_to {
        special = "Wildcard"
      }
      expression {
        contractual {
          schema_property_key = "custom_attribute_enum"
        }
        comparator = "Equals"
      }
    }

    inputs {
      defaults_to {
        value = "5"
      }
      expression {
        contractual {
          schema_property_key = "custom_attribute_integer"
        }
        comparator = "GreaterThan"
      }
    }

    inputs {
      defaults_to {
        value = "1000.0"
      }
      expression {
        contractual {
          schema_property_key = "custom_attribute_number"
        }
        comparator = "LessThanOrEquals"
      }
    }

    inputs {
      defaults_to {
        value = "true"
      }
      expression {
        contractual {
          schema_property_key = "custom_attribute_boolean"
        }
        comparator = "Equals"
      }
    }

    inputs {
      defaults_to {
        special = "CurrentTime"
      }
      expression {
        contractual {
          schema_property_key = "custom_attribute_date"
        }
        comparator = "GreaterThanOrEquals"
      }
    }

    inputs {
      defaults_to {
        special = "CurrentTime"
      }
      expression {
        contractual {
          schema_property_key = "custom_attribute_date"
        }
        comparator = "LessThanOrEquals"
      }
    }

    inputs {
      defaults_to {
        special = "CurrentTime"
      }
      expression {
        contractual {
          schema_property_key = "custom_attribute_datetime"
        }
        comparator = "NotEquals"
      }
    }

    inputs {
      defaults_to {
        special = "Wildcard"
      }
      expression {
        contractual {
          schema_property_key = "custom_attribute_for_empty_literal_block"
        }
        comparator = "Equals"
      }
    }

    inputs {
      defaults_to {
        special = "Wildcard"
      }
      expression {
        contractual {
          schema_property_key = "custom_attribute_for_empty_literal_value_type"
        }
        comparator = "Equals"
      }
    }

    inputs {
      defaults_to {
        value = genesyscloud_routing_queue.example_queue.id
      }
      expression {
        contractual {
          schema_property_key = "custom_attribute_queue"
          contractual {
            schema_property_key = "queue"
            contractual {
              schema_property_key = "id"
            }
          }
        }
        comparator = "Equals"
      }
    }

    inputs {
      defaults_to {
        values = ["general", "support"]
      }
      expression {
        contractual {
          schema_property_key = "custom_attribute_string_list"
        }
        comparator = "ContainsAny"
      }
    }

    outputs {
      defaults_to {
        value = genesyscloud_routing_queue.example_queue2.id
      }
      value {
        schema_property_key = "custom_attribute_queue"
        properties {
          schema_property_key = "queue"
          properties {
            schema_property_key = "id"
          }
        }
      }
    }

    outputs {
      defaults_to {
        value = "Premium Support"
      }
      value {
        schema_property_key = "custom_attribute_string"
      }
    }

    outputs {
      defaults_to {
        special = "Null"
      }
      value {
        schema_property_key = "custom_attribute_enum"
      }
    }

    outputs {
      defaults_to {
        values = ["basic_support", "general_help"]
      }
      value {
        schema_property_key = "custom_attribute_output_list"
      }
    }
  }

  rows {
    inputs {
      literal {
        value = "John Doe"
        type  = "string"
      }
    }
    inputs {
      literal {
        value = "option_1"
        type  = "string"
      }
    }
    inputs {
      literal {
        value = "85"
        type  = "integer"
      }
    }
    inputs {
      literal {
        value = "15000.0"
        type  = "number"
      }
    }
    inputs {
      literal {
        value = "true"
        type  = "boolean"
      }
    }
    inputs {
      literal {
        value = "2023-01-01"
        type  = "date"
      }
    }
    inputs {
      literal {
        value = "2023-12-31"
        type  = "date"
      }
    }
    inputs {
      literal {
        value = "2023-12-01T10:30:00.000Z"
        type  = "datetime"
      }
    }

    inputs {
      literal {} // Empty literal block with no value or type specified to use column default and must be provided
    }

    inputs {
      literal { // Literal block with empty value and type to use column default and must be provided
        value = ""
        type  = ""
      }
    }
    inputs {
      literal {
        value = genesyscloud_routing_queue.example_queue.id
        type  = "string"
      }
    }
    inputs {
      literal {
        value = "vip,premium"
        type  = "stringList"
      }
    }
    outputs {
      literal {
        value = genesyscloud_routing_queue.example_queue2.id
        type  = "string"
      }
    }
    outputs {
      literal {
        value = "Premium Support"
        type  = "string"
      }
    }
    outputs {
      literal {
        value = "option_2"
        type  = "string"
      }
    }
    outputs {
      literal {
        value = "premium_support,escalation,technical_expert" # Comma-separated string for stringList
        type  = "stringList"
      }
    }
  }
}
