locals {
  working_dir = {
    flow = "."
  }
  dependencies = {
    resource = [
      "../../data-sources/genesyscloud_auth_division_home/data-source.tf",
      "../genesyscloud_outbound_contact_list/resource.tf",
      "../genesyscloud_routing_wrapupcode/resource.tf",
    ]
  }
}
