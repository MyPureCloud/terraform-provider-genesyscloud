locals {
  dependencies = {
    resource = [
      "../genesyscloud_journey_views/resource.tf",
    ]
  }
  skip_if = {
    products_missing_any = ["journeyManagement", "cloudCX4"]
  }
}
