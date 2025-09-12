resource "genesyscloud_business_rules_decision_table" "example_decision_table" {
  name        = "Example Decision Table"
  description = "Example Decision Table created by terraform"
  division_id = data.genesyscloud_auth_division_home.home.id
  schema_id   = data.genesyscloud_business_rules_schema.comprehensive_schema.id

  columns {

    inputs {
      defaults_to {
        special = "Wildcard"
      }
      expression {
        contractual {
          schema_property_key = "customer_name"
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
          schema_property_key = "customer_type"
        }
        comparator = "NotEquals"
      }
    }

    inputs {
      defaults_to {
        value = "5"
      }
      expression {
        contractual {
          schema_property_key = "priority_score"
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
          schema_property_key = "revenue_amount"
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
          schema_property_key = "is_premium_member"
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
          schema_property_key = "customer_since"
        }
        comparator = "GreaterThan"
      }
    }

    inputs {
      defaults_to {
        special = "Null"
      }
      expression {
        contractual {
          schema_property_key = "last_interaction"
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
          schema_property_key = "current_queue"
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
        schema_property_key = "transfer_queue"
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
        schema_property_key = "assigned_skill"
      }
    }

    outputs {
      defaults_to {
        special = "Null"
      }
      value {
        schema_property_key = "escalation_level"
      }
    }
  }
}
