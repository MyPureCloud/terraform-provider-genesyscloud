resource "genesyscloud_outbound_campaignrule" "campaign_rule" {
  name                 = "Campaign Rule X"
  enabled              = false
  match_any_conditions = false
  campaign_rule_entities {
    campaign_ids = [genesyscloud_outbound_campaign.campaign.id]
    sequence_ids = [genesyscloud_outbound_sequence.example_outbound_sequence.id]
  }
  campaign_rule_conditions {
    condition_type = "campaignProgress"
    parameters {
      operator     = "lessThan"
      value        = "0.5"
      dialing_mode = "preview"
      priority     = "2"
    }
  }
  campaign_rule_actions {
    action_type = "turnOnCampaign"
    campaign_rule_action_entities {
      campaign_ids          = [genesyscloud_outbound_campaign.campaign2.id]
      use_triggering_entity = false
    }
    parameters {
      operator     = "lessThan"
      value        = "0.4"
      dialing_mode = "preview"
      priority     = "2"
    }
  }
}
