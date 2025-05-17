locals {
  dependencies = {
    resource = [
      "../genesyscloud_telephony_providers_edges_trunkbasesettings/resource.tf",
    ]
  }
  skip_if = {
    products_existing_any = ["hybridMedia"]
  }
}
