resource "genesyscloud_routing_queue" "example_queue" {
  name                              = "Example Queue"
  division_id                       = data.genesyscloud_auth_division_home.home.id
  description                       = "This is an example description"
  acw_wrapup_prompt                 = "MANDATORY_TIMEOUT"
  acw_timeout_ms                    = 300000
  skill_evaluation_method           = "BEST"
  queue_flow_id                     = data.genesyscloud_flow.default_inqueue_flow.id
  whisper_prompt_id                 = genesyscloud_architect_user_prompt.welcome_greeting.id
  auto_answer_only                  = true
  enable_transcription              = true
  enable_audio_monitoring           = true
  enable_manual_assignment          = true
  calling_party_name                = "Example Inc."
  outbound_messaging_sms_address_id = genesyscloud_routing_sms_address.example_routing_sms_address.id
  # outbound_email_address {
  #   domain_id = genesyscloud_routing_email_domain.main.id
  #   route_id  = genesyscloud_routing_email_route.support.id
  # }
  media_settings_call {
    alerting_timeout_sec      = 30
    service_level_percentage  = 0.7
    service_level_duration_ms = 10000
  }
  routing_rules {
    operator     = "MEETS_THRESHOLD"
    threshold    = 9
    wait_seconds = 300
  }
  bullseye_rings {
    expansion_timeout_seconds = 15.1
    skills_to_remove          = [genesyscloud_routing_skill.example_skill.id]

    member_groups {
      member_group_id   = genesyscloud_group.example_group.id
      member_group_type = "GROUP"
    }
  }
  default_script_ids = {
    EMAIL = genesyscloud_script.email.id
    # CHAT  = data.genesyscloud_script.chat.id
  }
  members {
    user_id  = genesyscloud_user.example_user.id
    ring_num = 2
  }
  wrapup_codes = [genesyscloud_routing_wrapupcode.win.id]
}
