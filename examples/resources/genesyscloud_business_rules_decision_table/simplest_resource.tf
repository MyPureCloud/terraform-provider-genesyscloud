
resource "genesyscloud_business_rules_decision_table" "simplest_decision_table" {
  name        = "Simplest Decision Table"
  description = "Minimal configuration example using rows_csv_filepath"
  division_id = data.genesyscloud_auth_division_home.home.id
  schema_id   = genesyscloud_business_rules_schema.example_business_rules_schema.id

  columns {
    inputs {
      defaults_to {
        value = "option_2"
      }
      expression {
        contractual {
          schema_property_key = "custom_attribute_enum"
        }
        comparator = "Equals"
      }
    }

    outputs {
      defaults_to {
        value = "anything"
      }
      value {
        schema_property_key = "custom_attribute_string"
      }
    }
  }

  # Prefer CSV for rows (nested rows { } is deprecated). Do not include a rowId column.
  rows_csv_filepath = "${path.module}/rows/example.csv"
}
