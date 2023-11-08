# When exporting outbound campaigns, any campaign with campaign_status = "stopping" will be
# given 5 minutes to stop. If it does not stop in that time it will not be exported.
# This is to avoid the exporter crashing when trying to export a campaign in the stopping state
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