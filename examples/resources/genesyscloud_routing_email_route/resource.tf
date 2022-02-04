resource "genesyscloud_routing_email_route" "support-route" {
  domain_id    = "test.example.com"
  pattern      = "support"
  from_name    = "Test Support"
  from_email   = "testsupport@test.example.com"
  queue_id     = genesyscloud_routing_queue.test-queue.id
  priority     = 5
  skill_ids    = [genesyscloud_routing_skill.support.id]
  language_id  = genesyscloud_routing_language.english.id
  flow_id      = data.genesyscloud_flow.flow.id
  spam_flow_id = data.genesyscloud_flow.spam_flow.id
  reply_email_address {
    domain_id = "test.example.com"
    route_id  = genesyscloud_routing_email_route.test.id
  }
  auto_bcc {
    name  = "Test Support"
    email = "support@test.example.com"
  }
}