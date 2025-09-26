resource "genesyscloud_business_rules_decision_table" "example_decision_table" {
  name        = "Example Decision Table"
  description = "Example Decision Table created by terraform"
  division_id = data.genesyscloud_auth_division_home.home.id
  schema_id   = genesyscloud_business_rules_schema.example_business_rules_schema.id

  columns {
    inputs {
      expression {
        contractual {
          schema_property_key = "custom_attribute_string"
        }
        comparator = "Equals"
      }
    }

    inputs {
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
        special = "Null"
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
        value = data.genesyscloud_routing_queue.standard_queue.id
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

    outputs {
      defaults_to {
        value = data.genesyscloud_routing_queue.vip_queue.id
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
      value {
        schema_property_key = "custom_attribute_enum"
      }
    }
  }

  rows {
    inputs {
      schema_property_key = "custom_attribute_string"
      literal {
        value = "John Doe"
        type  = "string"
      }
    }
    inputs {
      schema_property_key = "custom_attribute_enum"
      literal {
        value = "option_1"
        type  = "string"
      }
    }
    inputs {
      schema_property_key = "custom_attribute_integer"
      literal {
        value = "85"
        type  = "integer"
      }
    }
    inputs {
      schema_property_key = "custom_attribute_number"
      literal {
        value = "15000.0"
        type  = "number"
      }
    }
    inputs {
      schema_property_key = "custom_attribute_boolean"
      literal {
        value = "true"
        type  = "boolean"
      }
    }
    inputs {
      schema_property_key = "custom_attribute_date"
      comparator          = "GreaterThanOrEquals"
      literal {
        value = "2023-01-01"
        type  = "date"
      }
    }
    inputs {
      schema_property_key = "custom_attribute_date"
      comparator          = "LessThanOrEquals"
      literal {
        value = "2023-12-31"
        type  = "date"
      }
    }
    inputs {
      schema_property_key = "custom_attribute_datetime"
      literal {
        value = "2023-12-01T10:30:00.000Z"
        type  = "datetime"
      }
    }
    inputs {
      schema_property_key = "custom_attribute_queue"
      literal {
        value = data.genesyscloud_routing_queue.standard_queue.id
        type  = "string"
      }
    }
    outputs {
      schema_property_key = "custom_attribute_queue"
      literal {
        value = data.genesyscloud_routing_queue.vip_queue.id
        type  = "string"
      }
    }
    outputs {
      schema_property_key = "custom_attribute_string"
      literal {
        value = "Premium Support"
        type  = "string"
      }
    }
    outputs {
      schema_property_key = "custom_attribute_enum"
      literal {
        value = "option_2"
        type  = "string"
      }
    }
  }
}
