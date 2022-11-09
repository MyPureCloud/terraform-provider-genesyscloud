data "genesyscloud_routing_settings" "example-settings" {
  max_calls_per_agent                 = 10
  max_line_utilization                = 0.5
  abandon_seconds                     = 6.5
  compliance_abandon_rate_denominator = "ALL_CALLS"
}