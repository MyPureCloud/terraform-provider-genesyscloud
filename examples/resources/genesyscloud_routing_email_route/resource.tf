resource "genesyscloud_routing_email_route" "support-route" {
  domain_id    = "test.example.com"
  pattern      = "support"
  from_name    = "Test Support"
  from_email   = "testsupport@test.example.com"
  queue_id     = genesyscloud_routing_queue.test-queue.id
  priority     = 5
  skill_ids    = [genesyscloud_routing_skill.support.id]
  language_id  = genesyscloud_routing_language.english.id
  flow_id      = "34c17760-7539-11eb-9439-0242ac130002"
  spam_flow_id = "3fae0821-2a1a-4ebb-90b1-188b65923243"
  reply_email_address {
    domain_id = "test.example.com"
    route_id  = genesyscloud_routing_email_route.test.id
  }
  auto_bcc {
    name  = "Test Support"
    email = "support@test.example.com"
  }
}