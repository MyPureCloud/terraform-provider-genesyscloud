resource "genesyscloud_outbound_sequence" "example_outbound_sequence" {
  name         = "Example name"
  campaign_ids = [genesyscloud_outbound_campaign.campaign_1.id, genesyscloud_outbound_campaign.campaign_2.id]
  status       = "off"
  repeat       = false
}