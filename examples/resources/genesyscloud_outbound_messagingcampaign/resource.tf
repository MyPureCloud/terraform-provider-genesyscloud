resource "genesyscloud_outbound_messagingcampaign" "example_outbound_messagingcampaign" {
  name                 = "Example Messaging Campaign"
  division_id          = genesyscloud_auth_division.division.id
  campaign_status      = "off" // Possible values: on, off
  callable_time_set_id = genesyscloud_outbound_callabletimeset.callable_time_set.id
  contact_list_id      = genesyscloud_outbound_contact_list.contact_list.id
  dnc_list_ids         = [genesyscloud_outbound_dnclist.dnc_list_1.id, genesyscloud_outbound_dnclist.dnc_list_2.id]
  always_running       = true
  contact_sorts {
    field_name = "address"
    direction  = "ASC" // Possible values: ASC, DESC
    numeric    = false
  }
  messages_per_minute     = 10
  contact_list_filter_ids = [genesyscloud_outbound_contactlistfilter.contact_list_filter_1.id, genesyscloud_outbound_contactlistfilter.contact_list_filter_2.id]
  sms_config {
    message_column = "phone"
    phone_column   = "phone"
    sender_sms_phone_number {
      phone_number = "+123456789"
    }
    content_template_id = var.content_template_id
  }
}