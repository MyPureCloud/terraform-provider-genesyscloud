resource "genesyscloud_outbound_campaign" "campaign" {
  name                          = "Example Voice Campaign"
  dialing_mode                  = "agentless"
  caller_name                   = "John Doe"
  caller_address                = "+123456789"
  outbound_line_count           = 2
  campaign_status               = "off"
  contact_list_id               = genesyscloud_outbound_contact_list.contact_list.id
  site_id                       = genesyscloud_telephony_providers_edges_site.site.id
  call_analysis_response_set_id = genesyscloud_outbound_callanalysisresponseset.car.id
  phone_columns {
    column_name = "Cell"
  }
}