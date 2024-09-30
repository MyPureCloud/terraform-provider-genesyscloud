resource "genesyscloud_outbound_digitalruleset" "test_ruleset_sample" {
  name            = "Test Digital RuleSet - 13"
  version         = 1
  contact_list_id = "c2406f62-63e3-4826-a6f9-ae635bd731e0"
  rules {
    name     = "Rule-1"
    order    = 0
    category = "PreContact"
    conditions {
      inverted = true
      contact_column_condition_settings {
        column_name = "Work"
        operator    = "Equals"
        value       = "\"XYZ\""
        value_type  = "String"
      }
    }
    actions {
      do_not_send_action_settings = jsonencode({})
    }

  }
}