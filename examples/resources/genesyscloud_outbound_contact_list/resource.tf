resource "genesyscloud_outbound_contact_list" "contact_list" {
  name             = "Example Contact List"
  column_names     = ["First Name", "Last Name", "Cell", "Home", "Email", "Zipcode", "Timezone"]
  attempt_limit_id = genesyscloud_outbound_attempt_limit.attempt_limit.id

  phone_columns {
    column_name          = "Cell"
    type                 = "cell"
    callable_time_column = "Timezone"
  }
  phone_columns {
    column_name = "Home"
    type        = "home"
  }
  email_columns {
    column_name = "Email"
    type        = "work"
  }
}
