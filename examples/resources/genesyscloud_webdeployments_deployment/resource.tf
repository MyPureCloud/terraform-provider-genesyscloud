resource "genesyscloud_webdeployments_deployment" "example_deployment" {
  name            = "Example Web Deployment"
  description     = "This is an example of a web deployment"
  allowed_domains = ["example.com"]
  flow_id         = genesyscloud_flow.inbound_message_flow.id
  configuration {
    id      = genesyscloud_webdeployments_configuration.example_configuration.id
    version = genesyscloud_webdeployments_configuration.example_configuration.version
  }
}
