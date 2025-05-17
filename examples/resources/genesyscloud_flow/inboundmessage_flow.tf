resource "genesyscloud_flow" "inbound_message_flow" {
  filepath          = "${local.working_dir.flow}/inboundmessage_flow_example.yaml"
  file_content_hash = filesha256("${local.working_dir.flow}/inboundmessage_flow_example.yaml")
  substitutions = {
    flow_name          = "An example inbound message flow"
    home_division_name = data.genesyscloud_auth_division_home.home.name
  }
}
