resource "genesyscloud_outbound_contactlistfilter" "contact_list_filter" {
  name            = "Example CLF"
  contact_list_id = genesyscloud_outbound_contact_list.contact_list.id
  filter_type     = "OR"
  clauses {
    filter_type = "OR"
    predicates {
      column      = "Zipcode"
      column_type = "alphabetic"
      operator    = "EQUALS"
      value       = "ABC12345"
      inverted    = false
    }
  }
}