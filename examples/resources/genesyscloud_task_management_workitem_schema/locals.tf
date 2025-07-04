locals {
  dependencies = {
    resource = [
      "../../data-sources/genesyscloud_auth_division_home/data-source.tf"
    ]
  }
  skip_if = {
    products_missing_all = ["workitems"]
  }
}
