resource "genesyscloud_outbound_messagingcampaign" "example_outbound_messagingcampaign" {
  name                 = "Example Messaging Campaign"
  division_id          = data.genesyscloud_auth_division_home.home.id
  campaign_status      = "off" // Possible values: on, off
  callable_time_set_id = genesyscloud_outbound_callabletimeset.example_callable_time_set.id
  contact_list_id      = genesyscloud_outbound_contact_list.contact_list.id
  dnc_list_ids         = [genesyscloud_outbound_dnclist.dnc_list.id]
  always_running       = true
  contact_sorts {
    field_name = "Last Name"
    direction  = "ASC"
    numeric    = false
  }
  messages_per_minute     = 10
  contact_list_filter_ids = [genesyscloud_outbound_contactlistfilter.contact_list_filter.id]
  sms_config {
    phone_column            = "Cell"
    sender_sms_phone_number = local.sms_phone_number // "+123456789"
    content_template_id     = genesyscloud_responsemanagement_response.example_responsemanagement_response_sms.id
  }
}
