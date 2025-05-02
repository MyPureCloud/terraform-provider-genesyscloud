resource "genesyscloud_outbound_digitalruleset" "test_ruleset_sample" {
  name            = "Test Digital RuleSet - 13"
  contact_list_id = genesyscloud_outbound_contact_list.contact_list.id
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
