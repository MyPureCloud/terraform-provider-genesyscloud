resource "genesyscloud_flow" "Dave_-_SFDC_Service_-_v2_Basic" {
  file_content_hash = "${filesha256(var.genesyscloud_flow_Dave_-_SFDC_Service_-_v2_Basic_filepath)}"
  filepath          = "${var.genesyscloud_flow_Dave_-_SFDC_Service_-_v2_Basic_filepath}"
}

