locals {
  dependencies = {
    resource = [
      "../genesyscloud_flow/inboundmessage_flow.tf",
      "../genesyscloud_webdeployments_configuration/resource.tf",
    ]
  }
}
