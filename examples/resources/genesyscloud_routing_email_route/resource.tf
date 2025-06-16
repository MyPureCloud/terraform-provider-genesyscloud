resource "genesyscloud_routing_email_route" "example_route" {
  domain_id   = genesyscloud_routing_email_domain.example_domain_com.domain_id
  pattern     = "support"
  from_name   = "Support"
  from_email  = "support@${genesyscloud_routing_email_domain.example_domain_com.domain_id}"
  queue_id    = genesyscloud_routing_queue.example_queue.id
  priority    = 1
  skill_ids   = [genesyscloud_routing_skill.example_skill.id]
  language_id = genesyscloud_routing_language.english.id
  auto_bcc {
    name  = "Supervisors"
    email = "support_supervisors@${genesyscloud_routing_email_domain.example_domain_com.domain_id}"
  }
}

resource "genesyscloud_routing_email_route" "example_route_reference_other_route" {
  domain_id   = genesyscloud_routing_email_domain.example_domain_com.domain_id
  pattern     = "support_tier2"
  from_name   = "Example Support Tier 2"
  queue_id    = genesyscloud_routing_queue.example_queue.id
  priority    = 2
  skill_ids   = [genesyscloud_routing_skill.example_skill.id]
  language_id = genesyscloud_routing_language.english.id
  # flow_id      = genesyscloud_flow.inbound_call_flow.id
  # spam_flow_id = genesyscloud_flow.spam_call_flow.id
  reply_email_address {
    route_id  = genesyscloud_routing_email_route.example_route.id
    domain_id = genesyscloud_routing_email_domain.example_domain_com.domain_id
    # OR
    # self_reference_route = true
  }
}
