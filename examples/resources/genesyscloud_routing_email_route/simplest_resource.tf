resource "genesyscloud_routing_email_route" "example_route" {
  domain_id  = genesyscloud_routing_email_domain.example_domain_com.domain_id
  pattern    = "support"
  from_name  = "Support"
  from_email = "support@${genesyscloud_routing_email_domain.example_domain_com.domain_id}"
}
