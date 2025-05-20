locals {
  dependencies = {
    resource = [
      "../genesyscloud_outbound_campaign/resource.tf",
      "../genesyscloud_outbound_sequence/resource.tf",
    ]
  }
}
