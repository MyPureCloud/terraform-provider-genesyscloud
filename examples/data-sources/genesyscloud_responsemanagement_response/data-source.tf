data "genesyscloud_responsemanagement_response" "example_responsemanagement_response" {
  name    = "Responsemanagement response"
  library = genesyscloud_responsemanagement_library.library_1.id
}