locals {
  dependencies = [
    "../../data-sources/genesyscloud_auth_division_home/data-source.tf",
    "../genesyscloud_outbound_callabletimeset/resource.tf",
    "../genesyscloud_outbound_contact_list/resource.tf",
    "../genesyscloud_outbound_dnclist/resource.tf",
    "../genesyscloud_outbound_contactlistfilter/resource.tf",
    "../genesyscloud_responsemanagement_response/resource.tf",
  ]
  sms_phone_number = "+18159823725"
  // Constraints ensures this acceptance test example only gets run with certain conditions
  constraints = {

  }
}
