resource "genesyscloud_outbound_settings" "example_settings" {
  max_calls_per_agent                 = 10
  max_line_utilization                = 0.5
  abandon_seconds                     = 6.5
  compliance_abandon_rate_denominator = "ALL_CALLS"
  automatic_time_zone_mapping {
    callable_windows {
      mapped {
        earliest_callable_time = "09:00"
        latest_callable_time   = "17:00"
      }
      unmapped {
        earliest_callable_time = "08:00"
        latest_callable_time   = "18:00"
        time_zone_id           = "CET"
      }
    }
    supported_countries = ["US"]
  }
}
