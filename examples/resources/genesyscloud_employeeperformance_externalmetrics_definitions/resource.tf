resource "genesyscloud_employeeperformance_externalmetrics_definitions" "example_externalmetrics_definition" {
  name                   = "Example name"
  precision              = 2                // Between 0 and 5
  default_objective_type = "HigherIsBetter" // Possible values: HigherIsBetter, LowerIsBetter, TargetArea
  enabled                = true
  unit                   = "Currency" // Possible values: Seconds, Percent, Number, Currency
  unit_definition        = "USD"
}
