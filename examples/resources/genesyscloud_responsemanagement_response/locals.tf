locals {
  dependencies = {
    resource = [
      "../genesyscloud_responsemanagement_library/resource.tf",
      "../genesyscloud_responsemanagement_responseasset/resource.tf",
    ]
  }
}
