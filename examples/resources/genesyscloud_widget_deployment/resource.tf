resource "genesyscloud_widget_deployment" "mywidget" {
  name                    = "mywidget"
  description             = "My example widget test widget"
  flow_id                 = data.genesyscloud_flow.mytestflow.id
  client_type             = "v1"
  authentication_required = true
  disabled                = true
  client_config {
    authentication_url = "https://examplewebsite.com"
    webchat_skin       = "modern-caret-skin"
  }
  allowed_domains = []
}