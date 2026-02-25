# Example: Basic Business Unit
resource "genesyscloud_workforcemanagement_businessunits" "example_basic" {
  name = "Example Business Unit"
}

# Example: Business Unit with Settings
resource "genesyscloud_workforcemanagement_businessunits" "example_with_settings" {
  name        = "Example Business Unit with Settings"
  division_id = data.genesyscloud_auth_division_home.home.id

  settings {
    start_day_of_week = "Monday"
    time_zone         = "America/New_York"
  }
}

# Example: Business Unit with Short Term Forecasting
resource "genesyscloud_workforcemanagement_businessunits" "example_with_forecasting" {
  name = "Example Business Unit with Forecasting"

  settings {
    start_day_of_week = "Monday"
    time_zone         = "America/New_York"

    short_term_forecasting {
      default_history_weeks = 8
    }
  }
}

# Example: Business Unit with Scheduling Settings
resource "genesyscloud_workforcemanagement_businessunits" "example_with_scheduling" {
  name = "Example Business Unit with Scheduling"

  settings {
    start_day_of_week = "Monday"
    time_zone         = "America/New_York"

    scheduling {
      message_severities {
        type     = "AgentNotFound"
        severity = "Warning"
      }

      sync_time_off_properties = [
        "PayableMinutes"
      ]

      allow_work_plan_per_minute_granularity = false
    }
  }
}

# Example: Business Unit with Service Goal Impact Settings
resource "genesyscloud_workforcemanagement_businessunits" "example_with_service_goal_impact" {
  name = "Example Business Unit with Service Goal Impact"

  settings {
    start_day_of_week = "Monday"
    time_zone         = "America/New_York"

    scheduling {
      service_goal_impact {
        service_level {
          increase_by_percent = 10.0
          decrease_by_percent = 5.0
        }

        average_speed_of_answer {
          increase_by_percent = 15.0
          decrease_by_percent = 10.0
        }

        abandon_rate {
          increase_by_percent = 20.0
          decrease_by_percent = 15.0
        }
      }
    }
  }
}

# Example: Complete Business Unit Configuration
resource "genesyscloud_workforcemanagement_businessunits" "example_complete" {
  name        = "Example Complete Business Unit"
  division_id = data.genesyscloud_auth_division_home.home.id

  settings {
    start_day_of_week = "Monday"
    time_zone         = "America/New_York"

    short_term_forecasting {
      default_history_weeks = 8
    }

    scheduling {
      message_severities {
        type     = "AgentNotFound"
        severity = "Warning"
      }

      message_severities {
        type     = "UnableToProduceAgentSchedule"
        severity = "Error"
      }

      sync_time_off_properties = [
        "PayableMinutes"
      ]

      service_goal_impact {
        service_level {
          increase_by_percent = 10.0
          decrease_by_percent = 5.0
        }

        average_speed_of_answer {
          increase_by_percent = 15.0
          decrease_by_percent = 10.0
        }

        abandon_rate {
          increase_by_percent = 20.0
          decrease_by_percent = 15.0
        }
      }

      allow_work_plan_per_minute_granularity = true
    }
  }
}
