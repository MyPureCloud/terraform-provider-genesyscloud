locals {
  # TODO: Remove this skip constraint when the guide feature is fully released
  skip_if = {
    feature_toggles_required = ["guide"]
  }
} 