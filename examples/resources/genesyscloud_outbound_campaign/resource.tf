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
  call_analysis_response_set_id = genesyscloud_outbound_callanalysisresponseset.example_cars.id
  phone_columns {
    column_name = "Cell"
  }
}

# Example Power dialing campaign with diagnostics settings
# diagnostics_settings is only applicable to Power and Predictive dialing modes
resource "genesyscloud_outbound_campaign" "power_campaign" {
  name                          = "Example Power Campaign"
  dialing_mode                  = "power"
  caller_name                   = "Support Team"
  caller_address                = "+15559876543"
  outbound_line_count           = 5
  abandon_rate                  = 5.0
  max_calls_per_agent           = 2.0
  campaign_status               = "off"
  contact_list_id               = genesyscloud_outbound_contact_list.contact_list.id
  queue_id                      = genesyscloud_routing_queue.queue.id
  site_id                       = genesyscloud_telephony_providers_edges_site.site.id
  call_analysis_response_set_id = genesyscloud_outbound_callanalysisresponseset.example_cars.id
  phone_columns {
    column_name = "Cell"
  }
  # Diagnostics settings - controls campaign health alerts
  # report_low_max_calls_per_agent_alert: When true (default), generates a health alert
  # if Max Calls Per Agent is set below the value in Outbound Settings
  diagnostics_settings {
    report_low_max_calls_per_agent_alert = true
  }
}

resource "genesyscloud_outbound_campaign" "campaign2" {
  name                          = "Example Voice Campaign2"
  dialing_mode                  = "agentless"
  caller_name                   = "Jane Doe"
  caller_address                = "+15551234567"
  outbound_line_count           = 2
  campaign_status               = "off"
  contact_list_id               = genesyscloud_outbound_contact_list.contact_list.id
  site_id                       = genesyscloud_telephony_providers_edges_site.site.id
  call_analysis_response_set_id = genesyscloud_outbound_callanalysisresponseset.example_cars.id
  phone_columns {
    column_name = "Cell"
  }
}
