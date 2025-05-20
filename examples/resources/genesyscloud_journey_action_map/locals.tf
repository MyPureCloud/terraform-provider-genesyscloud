locals {
  dependencies = {
    resource = [
      "../genesyscloud_journey_segment/resource.tf",
      "../genesyscloud_flow/resource.tf",
    ]
  }
  skip_if = {
    products_missing_any = ["journeyManagement", "cloudCX4"]
  }

}
