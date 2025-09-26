locals {
  skip_if = {
    products_missing_all     = ["ruleBasedDecisions"]
    feature_toggles_required = ["PURE-5186_CoreRulesAndDecisions"]
  }
}
