resource "genesyscloud_outbound_wrapupcodemappings" "example_mappings" {
  default_set = ["Right_Party_Contact", "Contact_UnCallable"]
  mappings {
    wrapup_code_id = genesyscloud_routing_wrapupcode.unknown.id
    flags          = ["Contact_UnCallable", "Number_UnCallable", ]
  }
  mappings {
    wrapup_code_id = genesyscloud_routing_wrapupcode.win.id
    flags          = ["Right_Party_Contact"]
  }
}
