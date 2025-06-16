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
}
