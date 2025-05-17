locals {
  dependencies = {
    resource = [
      "../genesyscloud_journey_outcome/resource.tf",
    ]
  }
  skip_if = {
    products_missing_any = ["journeyManagement", "cloudCX4"]
  }
}
