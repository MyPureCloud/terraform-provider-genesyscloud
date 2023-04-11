resource "genesyscloud_outbound_contact_list" "contact-list" {
  name             = "Example Contact List"
  column_names     = ["First Name", "Last Name", "Cell", "Home"]
  attempt_limit_id = genesyscloud_outbound_attempt_limit.attempt-limit.id
  phone_columns {
    column_name = "Cell"
    type        = "cell"
  }
  phone_columns {
    column_name = "Home"
    type        = "home"
  }
}
