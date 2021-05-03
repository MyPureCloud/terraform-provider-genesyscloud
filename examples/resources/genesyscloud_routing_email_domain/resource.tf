resource "genesyscloud_routing_email_domain" "test-example-com" {
  domain_id             = "test.example.com"
  subdomain             = false
  mail_from_domain      = "example.com"
  custom_smtp_server_id = "99490182-2695-47db-a17d-0bf2ef230827"
}
