locals {
  dependencies = {
    resource = []
  }
  skip_if = {
    products_missing_any = ["journeyManagement", "cloudCX4"]
  }
}
