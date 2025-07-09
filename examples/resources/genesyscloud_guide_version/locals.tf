locals {
  dependencies = {
    resource = [
      "../genesyscloud_guide/resource.tf",
      "../genesyscloud_integration_action/resource.tf",
    ]
  }
  # TODO: Remove when the guide feature is fully released
  skip_if = {
    feature_toggles_required = ["guide"]
  }
}