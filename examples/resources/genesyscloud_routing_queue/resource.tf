resource "genesyscloud_routing_queue" "test_queue" {
  name                              = "Test Queue"
  division_id                       = "505e1036-6f04-405c-a630-de94a8ad2eb8"
  description                       = "This is a test queue"
  acw_wrapup_prompt                 = "MANDATORY_TIMEOUT"
  acw_timeout_ms                    = 300000
  skill_evaluation_method           = "BEST"
  queue_flow_id                     = "34c17760-7539-11eb-9439-0242ac130002"
  whisper_prompt_id                 = "3fae0821-2a1a-4ebb-90b1-188b65923243"
  auto_answer_only                  = true
  enable_transcription              = true
  enable_manual_assignment          = true
  calling_party_name                = "Acme Inc."
  outbound_messaging_sms_address_id = "c1bb045e-254d-4316-9d78-cea6849a3db4"
  outbound_email_address {
    domain_id = "example.com"
    route_id  = "1b242045-d0f9-49e0-b07f-de19fa4e374e"
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
    skills_to_remove          = ["0d9d3b90-53a8-43cf-a3ad-7e6a41592a3f"]
  }
  default_script_ids {
    EMAIL = "153fcff5-597e-4f17-94e5-17eac456a0b2"
    CHAT  = "98dff282-c50c-4c36-bc70-80b058564e1b"
  }
  members {
    user_id  = "851dcc4f-80d4-4cc9-8bb3-b98cf560d572"
    ring_num = 2
  }
}
