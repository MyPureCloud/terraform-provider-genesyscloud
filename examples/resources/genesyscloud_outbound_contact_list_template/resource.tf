resource "genesyscloud_outbound_contact_list_template" "contact_list_template" {
  name             = "Example Contact List Template"
  column_names     = ["First Name", "Last Name", "Cell", "Home"]
  attempt_limit_id = genesyscloud_outbound_attempt_limit.attempt_limit.id
  phone_columns {
    column_name = "Cell"
    type        = "cell"
  }
  phone_columns {
    column_name = "Home"
    type        = "home"
  }
}
