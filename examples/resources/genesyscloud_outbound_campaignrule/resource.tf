resource "genesyscloud_outbound_campaignrule" "campaign_rule" {
  name                 = "Campaign Rule X"
  enabled              = false
  match_any_conditions = false
  campaign_rule_entities {
    campaign_ids = ["e5e0a123-4321-123a-a122-10e55c123r11"]
    sequence_ids = ["f5e3a111-j2r4-13rr-pa12-52psc1fe5r29"]
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
      campaign_ids          = ["ebce858c-28d1-403f-9dd4-8d9997b95104"]
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