resource "genesyscloud_routing_queue" "test_queue" {
  name                              = "Test Queue"
  division_id                       = genesyscloud_auth_division.home.id
  description                       = "This is a test queue"
  acw_wrapup_prompt                 = "MANDATORY_TIMEOUT"
  acw_timeout_ms                    = 300000
  skill_evaluation_method           = "BEST"
  queue_flow_id                     = "34c17760-7539-11eb-9439-0242ac130002"
  whisper_prompt_id                 = "3fae0821-2a1a-4ebb-90b1-188b65923243"
  auto_answer_only                  = true
  enable_transcription              = true
  enable_manual_assignment          = true
  calling_party_name                = "Example Inc."
  outbound_messaging_sms_address_id = "c1bb045e-254d-4316-9d78-cea6849a3db4"
  outbound_email_address {
    domain_id = genesyscloud_routing_email_domain.main.id
    route_id  = genesyscloud_routing_email_route.support.id
  }
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
    skills_to_remove          = [genesyscloud_routing_skill.test-skill.id]
  }
  default_script_ids = {
    EMAIL = "153fcff5-597e-4f17-94e5-17eac456a0b2"
    CHAT  = "98dff282-c50c-4c36-bc70-80b058564e1b"
  }
  members {
    user_id  = genesyscloud_user.test-user.id
    ring_num = 2
  }
  wrapup_codes = [genesyscloud_routing_wrapupcode.test-code.id]
}
