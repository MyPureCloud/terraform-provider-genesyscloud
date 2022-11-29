resource "genesyscloud_employeeperformance_externalmetrics_definition" "example_externalmetrics_definition" {
  name                   = "Example name"
  precision              = 0                // Between 0 and 5
  default_objective_type = "HigherIsBetter" // Possible values: HigherIsBetter, LowerIsBetter, TargetArea
  enabled                = true
  unit                   = "Seconds" // Possible values: Seconds, Percent, Number, Currency
  unit_definition        = ""
}