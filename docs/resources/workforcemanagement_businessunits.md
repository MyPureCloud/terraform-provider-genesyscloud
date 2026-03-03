---
page_title: "genesyscloud_workforcemanagement_businessunits Resource - terraform-provider-genesyscloud"
subcategory: ""
description: |-
  Genesys Cloud workforce management business units
---
# genesyscloud_workforcemanagement_businessunits (Resource)

Genesys Cloud workforce management business units

## API Usage
The following Genesys Cloud APIs are used by this resource. Ensure your OAuth Client has been granted the necessary scopes and permissions to perform these operations:

* [POST /api/v2/workforcemanagement/businessunits](https://developer.mypurecloud.com/api/rest/v2/workforcemanagement/#post-api-v2-workforcemanagement-businessunits)
* [GET /api/v2/workforcemanagement/businessunits](https://developer.mypurecloud.com/api/rest/v2/workforcemanagement/#get-api-v2-workforcemanagement-businessunits)
* [GET /api/v2/workforcemanagement/businessunits/{businessUnitId}](https://developer.mypurecloud.com/api/rest/v2/workforcemanagement/#get-api-v2-workforcemanagement-businessunits--businessUnitId-)
* [PATCH /api/v2/workforcemanagement/businessunits/{businessUnitId}](https://developer.mypurecloud.com/api/rest/v2/workforcemanagement/#patch-api-v2-workforcemanagement-businessunits--businessUnitId-)
* [DELETE /api/v2/workforcemanagement/businessunits/{businessUnitId}](https://developer.mypurecloud.com/api/rest/v2/workforcemanagement/#delete-api-v2-workforcemanagement-businessunits--businessUnitId-)

## Permissions and Scopes

The following permissions are required to use this resource:

* `coaching:appointment:add`
* `coaching:appointment:edit`
* `learning:assignment:add`
* `learning:assignment:reschedule`
* `wfm:activityCode:add`
* `wfm:activityCode:delete`
* `wfm:activityCode:edit`
* `wfm:activityCode:view`
* `wfm:agent:edit`
* `wfm:agent:view`
* `wfm:agentSchedule:view`
* `wfm:agentShiftTradeRequest:participate`
* `wfm:agentTimeOffRequest:submit`
* `wfm:businessUnit:add`
* `wfm:businessUnit:delete`
* `wfm:businessUnit:edit`
* `wfm:businessUnit:view`
* `wfm:historicalAdherence:view`
* `wfm:intraday:view`
* `wfm:managementUnit:add`
* `wfm:managementUnit:delete`
* `wfm:managementUnit:edit`
* `wfm:managementUnit:view`
* `wfm:planningGroup:add`
* `wfm:planningGroup:delete`
* `wfm:planningGroup:edit`
* `wfm:planningGroup:view`
* `wfm:publishedSchedule:view`
* `wfm:realtimeAdherence:view`
* `wfm:schedule:add`
* `wfm:schedule:delete`
* `wfm:schedule:edit`
* `wfm:schedule:generate`
* `wfm:schedule:view`
* `wfm:serviceGoalTemplate:add`
* `wfm:serviceGoalTemplate:delete`
* `wfm:serviceGoalTemplate:edit`
* `wfm:serviceGoalTemplate:view`
* `wfm:shiftTradeRequest:edit`
* `wfm:shiftTradeRequest:view`
* `wfm:shortTermForecast:add`
* `wfm:shortTermForecast:delete`
* `wfm:shortTermForecast:edit`
* `wfm:shortTermForecast:view`
* `wfm:shrinkage:view`
* `wfm:staffingGroup:add`
* `wfm:staffingGroup:delete`
* `wfm:staffingGroup:edit`
* `wfm:staffingGroup:view`
* `wfm:timeOffLimit:add`
* `wfm:timeOffLimit:delete`
* `wfm:timeOffLimit:edit`
* `wfm:timeOffLimit:view`
* `wfm:timeOffPlan:add`
* `wfm:timeOffPlan:delete`
* `wfm:timeOffPlan:edit`
* `wfm:timeOffPlan:view`
* `wfm:timeOffRequest:add`
* `wfm:timeOffRequest:edit`
* `wfm:timeOffRequest:view`
* `wfm:workPlan:add`
* `wfm:workPlan:delete`
* `wfm:workPlan:edit`
* `wfm:workPlan:view`
* `wfm:workPlanRotation:add`
* `wfm:workPlanRotation:delete`
* `wfm:workPlanRotation:edit`
* `wfm:workPlanRotation:view`

The following OAuth scopes are required to use this resource:

* `coaching`
* `learning`
* `workforce-management`
* `workforce-management:readonly`


## Example Usage

```terraform
# Example: Basic Business Unit
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The name of the business unit
- `settings` (Block List, Min: 1, Max: 1) Configuration for the business unit (see [below for nested schema](#nestedblock--settings))

### Optional

- `division_id` (String) The ID of the division to which the business unit should be added. If not set the home division will be used

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--settings"></a>
### Nested Schema for `settings`

Required:

- `start_day_of_week` (String) The start day of week for this business unit
- `time_zone` (String) The time zone for this business unit, using the Olsen tz database format

Optional:

- `scheduling` (Block List, Max: 1) Scheduling settings (see [below for nested schema](#nestedblock--settings--scheduling))
- `short_term_forecasting` (Block List, Max: 1) Short term forecasting settings (see [below for nested schema](#nestedblock--settings--short_term_forecasting))

Read-Only:

- `metadata` (List of Object) Version metadata for this business unit (see [below for nested schema](#nestedatt--settings--metadata))

<a id="nestedblock--settings--scheduling"></a>
### Nested Schema for `settings.scheduling`

Optional:

- `allow_work_plan_per_minute_granularity` (Boolean) Indicates whether or not per minute granularity for scheduling will be enabled for this business unit. Defaults to false.
- `message_severities` (Block List) Schedule generation message severity configuration (see [below for nested schema](#nestedblock--settings--scheduling--message_severities))
- `service_goal_impact` (Block List, Max: 1) Configures the max percent increase and decrease of service goals for this business unit (see [below for nested schema](#nestedblock--settings--scheduling--service_goal_impact))
- `sync_time_off_properties` (List of String) Synchronize set of time off properties from scheduled activities to time off requests when the schedule is published.

<a id="nestedblock--settings--scheduling--message_severities"></a>
### Nested Schema for `settings.scheduling.message_severities`

Optional:

- `severity` (String) The severity of the message. Validation is handled by the API to avoid maintaining a potentially stale list of enum values. See API documentation for valid values: https://developer.genesys.cloud/useragentman/workforcemanagement/#post-api-v2-workforcemanagement-businessunits
- `type` (String) The type of the message. Validation is handled by the API to avoid maintaining a potentially stale list of enum values. See API documentation for valid values: https://developer.genesys.cloud/useragentman/workforcemanagement/#post-api-v2-workforcemanagement-businessunits


<a id="nestedblock--settings--scheduling--service_goal_impact"></a>
### Nested Schema for `settings.scheduling.service_goal_impact`

Required:

- `abandon_rate` (Block List, Min: 1, Max: 1) Allowed abandon rate percent increase and decrease (see [below for nested schema](#nestedblock--settings--scheduling--service_goal_impact--abandon_rate))
- `average_speed_of_answer` (Block List, Min: 1, Max: 1) Allowed average speed of answer percent increase and decrease (see [below for nested schema](#nestedblock--settings--scheduling--service_goal_impact--average_speed_of_answer))
- `service_level` (Block List, Min: 1, Max: 1) Allowed service level percent increase and decrease (see [below for nested schema](#nestedblock--settings--scheduling--service_goal_impact--service_level))

<a id="nestedblock--settings--scheduling--service_goal_impact--abandon_rate"></a>
### Nested Schema for `settings.scheduling.service_goal_impact.abandon_rate`

Required:

- `decrease_by_percent` (Number) The maximum allowed percent decrease from the configured goal
- `increase_by_percent` (Number) The maximum allowed percent increase from the configured goal


<a id="nestedblock--settings--scheduling--service_goal_impact--average_speed_of_answer"></a>
### Nested Schema for `settings.scheduling.service_goal_impact.average_speed_of_answer`

Required:

- `decrease_by_percent` (Number) The maximum allowed percent decrease from the configured goal
- `increase_by_percent` (Number) The maximum allowed percent increase from the configured goal


<a id="nestedblock--settings--scheduling--service_goal_impact--service_level"></a>
### Nested Schema for `settings.scheduling.service_goal_impact.service_level`

Required:

- `decrease_by_percent` (Number) The maximum allowed percent decrease from the configured goal
- `increase_by_percent` (Number) The maximum allowed percent increase from the configured goal




<a id="nestedblock--settings--short_term_forecasting"></a>
### Nested Schema for `settings.short_term_forecasting`

Optional:

- `default_history_weeks` (Number) The number of historical weeks to consider when creating a forecast. This setting is only used for legacy weighted average forecasts


<a id="nestedatt--settings--metadata"></a>
### Nested Schema for `settings.metadata`

Read-Only:

- `version` (Number)

