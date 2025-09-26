
resource "genesyscloud_business_rules_decision_table" "simplest_decision_table" {
  name        = "Simplest Decision Table"
  description = "Minimal configuration example"
  division_id = data.genesyscloud_auth_division_home.home.id
  schema_id   = genesyscloud_business_rules_schema.example_business_rules_schema.id

  columns {
    inputs {
      expression {
        contractual {
          schema_property_key = "custom_attribute_enum"
        }
        comparator = "Equals"
      }
    }

    outputs {
      value {
        schema_property_key = "custom_attribute_string"
      }
    }
  }

  rows {
    inputs {
      schema_property_key = "custom_attribute_enum"
      literal {
        value = "option_1"
        type  = "string"
      }
    }
    outputs {
      schema_property_key = "custom_attribute_string"
      literal {
        value = "high"
        type  = "string"
      }
    }
  }
}
