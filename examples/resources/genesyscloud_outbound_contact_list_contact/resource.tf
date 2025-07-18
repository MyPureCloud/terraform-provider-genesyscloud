resource "genesyscloud_outbound_contact_list_contact" "contact" {
  contact_list_id = genesyscloud_outbound_contact_list.contact_list.id
  callable        = true
  data = {
    "First Name" = "John"
    "Last Name"  = "Doe"
    Cell         = "+1111111"
    Home         = "+2222222"
    Email        = "example@email.com"
    Zipcode      = "12345"
    Timezone     = "EST"
  }
  phone_number_status {
    key      = "Cell"
    callable = true
  }
  phone_number_status {
    key      = "Home"
    callable = false
  }
  contactable_status {
    media_type  = "Voice"
    contactable = true
    column_status {
      column      = "Cell"
      contactable = true
    }
    column_status {
      column      = "Home"
      contactable = false
    }
  }
  contactable_status {
    media_type  = "Email"
    contactable = true
    column_status {
      column      = "Email"
      contactable = true
    }
  }

  contactable_status {
    contactable = false
    media_type  = "WhatsApp"
  }
}
