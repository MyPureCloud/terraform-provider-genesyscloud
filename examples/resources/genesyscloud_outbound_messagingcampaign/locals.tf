locals {
  dependencies = {
    resource = [
      "../../data-sources/genesyscloud_auth_division_home/data-source.tf",
      "../genesyscloud_outbound_callabletimeset/resource.tf",
      "../genesyscloud_outbound_contact_list/resource.tf",
      "../genesyscloud_outbound_dnclist/resource.tf",
      "../genesyscloud_outbound_contactlistfilter/resource.tf",
      "../genesyscloud_responsemanagement_response/resource.tf",
    ]
    simplest_resource = [
      "../../data-sources/genesyscloud_auth_division_home/data-source.tf",
      "../genesyscloud_outbound_callabletimeset/resource.tf",
      "../genesyscloud_outbound_contact_list/resource.tf",
      "../genesyscloud_outbound_dnclist/resource.tf",
      "../genesyscloud_outbound_contactlistfilter/resource.tf",
    ]
  }
  sms_phone_number = "+18159823725"
  skip_if = {
    not_in_domains = ["inintca.com"]
  }
}
