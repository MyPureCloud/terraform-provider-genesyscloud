resource "genesyscloud_outbound_wrapupcodemappings" "example_mappings" {
  default_set = ["Right_Party_Contact", "Contact_UnCallable"]
  mappings {
    wrapup_code_id = genesyscloud_routing_wrapupcode.wrapup_code_1.id
    flags          = ["Contact_UnCallable"]
  }
  mappings {
    wrapup_code_id = genesyscloud_routing_wrapupcode.wrapup_code_2.id
    flags          = ["Number_UnCallable", "Right_Party_Contact"]
  }
}  