resource "genesyscloud_routing_email_domain" "example_domain_com" {
  domain_id        = "example.com"
  subdomain        = false
  mail_from_domain = "mail.example.com"
}
