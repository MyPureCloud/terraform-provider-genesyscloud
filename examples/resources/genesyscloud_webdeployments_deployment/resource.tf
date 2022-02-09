data "genesyscloud_flow" "incomingMessageFlow" {
  name = "Incoming Message Flow"
}

data "genesyscloud_webdeployments_configuration" "exampleConfiguration" {
  name = "Example Web Deployment Configuration"
}

resource "genesyscloud_webdeployments_deployment" "exampleDeployment" {
  name            = "Example Web Deployment"
  description     = "This is an example of a web deployment"
  allowed_domains = ["genesys.com"]
  flow_id         = data.genesyscloud_flow.incomingMessageFlow.id
  configuration {
    id      = data.genesyscloud_webdeployments_configuration.exampleConfiguration.id
    version = data.genesyscloud_webdeployments_configuration.exampleConfiguration.version
  }
}