resource "genesyscloud_outbound_ruleset" "example_outbound_ruleset" {
  name            = "Example Ruleset"
  contact_list_id = genesyscloud_outbound_contact_list.contact_list.id
  queue_id        = genesyscloud_routing_queue.example_queue.id

  rules {
    name     = "Do Not Attempt To Dial Contacts Without Phone Number"
    order    = 0
    category = "DIALER_PRECALL"
    conditions {
      type           = "contactAttributeCondition"
      inverted       = true
      attribute_name = "Phone"
      value          = ""
      value_type     = "STRING"
      operator       = "EQUALS"
    }
    actions {
      type             = "Action"
      action_type_name = "DO_NOT_DIAL"
      properties       = {}
    }
  }

  rules {
    name     = "When Call Is A Win, Call Data Action"
    order    = 1
    category = "DIALER_WRAPUP" // Possible values: DIALER_PRECALL, DIALER_WRAPUP
    conditions {
      type     = "wrapupCondition" // Possible values: wrapupCondition, systemDispositionCondition, contactAttributeCondition, phoneNumberCondition, phoneNumberTypeCondition, callAnalysisCondition, contactPropertyCondition, dataActionCondition
      inverted = false
      codes    = [genesyscloud_routing_wrapupcode.win.id]
    }
    actions {
      type                       = "dataActionBehavior"
      action_type_name           = "DATA_ACTION"
      properties                 = {}
      data_action_id             = genesyscloud_integration_action.example_action.id
      call_analysis_result_field = "examplestr"
    }
  }

  rules {
    name     = "Retry on Busy"
    order    = 2
    category = "DIALER_WRAPUP"
    conditions {
      type     = "callAnalysisCondition"
      inverted = false
      value    = "disposition.classification.callable.busy"
      operator = "EQUALS"
    }
    actions {
      type             = "Action"
      action_type_name = "SCHEDULE_CALLBACK"
      properties = {
        callbackOffset = "5"
      }
    }
  }

  rules {
    name     = "Designated ID for 555 area code"
    order    = 3
    category = "DIALER_PRECALL"
    conditions {
      type     = "phoneNumberCondition"
      inverted = false
      value    = "555"
      operator = "BEGINS_WITH"
    }
    actions {
      type             = "Action"
      action_type_name = "SET_CALLER_ID"
      properties = {
        callerAddress = "+18001234567"
        callerName    = "Acme Inc"
      }
    }
  }
}
