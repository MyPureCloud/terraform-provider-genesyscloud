resource "genesyscloud_widget_deployment" "mywidget" {
  name                    = "mywidget"
  description             = "My example widget test widget"
  flow_id                 = data.genesyscloud_flow.mytestflow.id
  client_type             = "v1"
  authentication_required = true
  disabled                = true
  third_party_client_config = {
    foo = "bar"
  }
  allowed_domains = []
}