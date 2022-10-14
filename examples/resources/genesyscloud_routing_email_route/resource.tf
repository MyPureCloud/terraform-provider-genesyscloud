resource "genesyscloud_routing_email_route" "support-route" {
  domain_id    = "example.domain.com"
  pattern      = "support"
  from_name    = "Example Support"
  from_email   = "examplesupport@example.domain.com"
  queue_id     = genesyscloud_routing_queue.example-queue.id
  priority     = 5
  skill_ids    = [genesyscloud_routing_skill.support.id]
  language_id  = genesyscloud_routing_language.english.id
  flow_id      = data.genesyscloud_flow.flow.id
  spam_flow_id = data.genesyscloud_flow.spam_flow.id
  reply_email_address {
    domain_id = "example.domain.com"
    route_id  = genesyscloud_routing_email_route.example.id
  }
  auto_bcc {
    name  = "Example Support"
    email = "support@example.domain.com"
  }
}
