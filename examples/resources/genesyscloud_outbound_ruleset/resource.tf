resource "genesyscloud_outbound_ruleset" "example_outbound_ruleset" {
  name            = ""
  contact_list_id = genesyscloud_outbound_contact_list.contact_list.id
  queue_id        = genesyscloud_routing_queue.queue.id
  rules {
    name     = ""
    order    = 0
    category = "DIALER_PRECALL" // Possible values: DIALER_PRECALL, DIALER_WRAPUP
    conditions {
      type                       = "wrapupCondition" // Possible values: wrapupCondition, systemDispositionCondition, contactAttributeCondition, phoneNumberCondition, phoneNumberTypeCondition, callAnalysisCondition, contactPropertyCondition, dataActionCondition
      inverted                   = true
      attribute_name             = ""
      value                      = ""
      value_type                 = "STRING" // Possible values: STRING, NUMERIC, DATETIME, PERIOD
      operator                   = "EQUALS" // Possible values: EQUALS, LESS_THAN, LESS_THAN_EQUALS, GREATER_THAN, GREATER_THAN_EQUALS, CONTAINS, BEGINS_WITH, ENDS_WITH, BEFORE, AFTER, IN
      codes                      = []
      property                   = ""
      property_type              = "LAST_ATTEMPT_BY_COLUMN" // Possible values: LAST_ATTEMPT_BY_COLUMN, LAST_ATTEMPT_OVERALL, LAST_WRAPUP_BY_COLUMN, LAST_WRAPUP_OVERALL
      data_action_id             = genesyscloud_integration_action.data_action.id
      data_not_found_resolution  = true
      contact_id_field           = ""
      call_analysis_result_field = ""
      agent_wrapup_field         = ""
      contact_column_to_data_action_field_mappings {
        contact_column_name = ""
        data_action_field   = ""
      }
      predicates {
        output_field                    = ""
        output_operator                 = "EQUALS" // Possible values: EQUALS, LESS_THAN, LESS_THAN_EQUALS, GREATER_THAN, GREATER_THAN_EQUALS, CONTAINS, BEGINS_WITH, ENDS_WITH, BEFORE, AFTER
        comparison_value                = ""
        inverted                        = true
        output_field_missing_resolution = true
      }
    }
    actions {
      type             = "Action"      // Possible values: Action, modifyContactAttribute, dataActionBehavior
      action_type_name = "DO_NOT_DIAL" // Possible values: DO_NOT_DIAL, MODIFY_CONTACT_ATTRIBUTE, SWITCH_TO_PREVIEW, APPEND_NUMBER_TO_DNC_LIST, SCHEDULE_CALLBACK, CONTACT_UNCALLABLE, NUMBER_UNCALLABLE, SET_CALLER_ID, SET_SKILLS, DATA_ACTION
      update_option    = "SET"         // Possible values: SET, INCREMENT, DECREMENT, CURRENT_TIME
      properties       = {}
      data_action_id   = genesyscloud_integration_action.data_action.id
      contact_column_to_data_action_field_mappings {
        contact_column_name = ""
        data_action_field   = ""
      }
      contact_id_field           = ""
      call_analysis_result_field = ""
      agent_wrapup_field         = ""
    }
  }
}