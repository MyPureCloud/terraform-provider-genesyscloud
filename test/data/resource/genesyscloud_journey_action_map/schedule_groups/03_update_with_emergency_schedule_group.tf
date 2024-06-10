resource "genesyscloud_journey_action_map" "terraform_test_-TEST-CASE-" {
  display_name          = "terraform_test_-TEST-CASE-_updated"
  trigger_with_segments = [genesyscloud_journey_segment.terraform_test_-TEST-CASE-_action_map_dependency.id]
  activation {
    type = "immediate"
  }
  action {
    media_type = "webMessagingOffer"
  }
  action_map_schedule_groups {
    action_map_schedule_group_id           = genesyscloud_architect_schedulegroups.terraform_test_-TEST-CASE-_action_map_dependency_closed.id
    emergency_action_map_schedule_group_id = genesyscloud_architect_schedulegroups.terraform_test_-TEST-CASE-_action_map_dependency_open.id
  }
  start_date = "2022-07-04T12:00:00.000000"

  depends_on = [
    genesyscloud_journey_segment.terraform_test_-TEST-CASE-_action_map_dependency,
    genesyscloud_architect_schedulegroups.terraform_test_-TEST-CASE-_action_map_dependency_open,
    genesyscloud_architect_schedulegroups.terraform_test_-TEST-CASE-_action_map_dependency_closed
  ]
}

resource "genesyscloud_journey_segment" "terraform_test_-TEST-CASE-_action_map_dependency" {
  display_name            = "terraform_test_-TEST-CASE-_action_map_dependency"
  color                   = "#008000"
  scope                   = "Session"
  should_display_to_agent = false
  journey {
    patterns {
      criteria {
        key                = "page.title"
        values             = ["Title"]
        operator           = "notEqual"
        should_ignore_case = true
      }
      count        = 1
      stream_type  = "Web"
      session_type = "web"
    }
  }
}

resource "genesyscloud_architect_schedulegroups" "terraform_test_-TEST-CASE-_action_map_dependency_open" {
  name              = "terraform_test_-TEST-CASE-_action_map_dependency_open"
  division_id       = null
  time_zone         = "Asia/Singapore"
  open_schedules_id = [genesyscloud_architect_schedules.terraform_test_-TEST-CASE-_action_map_dependency_open.id]

  depends_on = [genesyscloud_architect_schedules.terraform_test_-TEST-CASE-_action_map_dependency_open]
}

resource "genesyscloud_architect_schedulegroups" "terraform_test_-TEST-CASE-_action_map_dependency_closed" {
  name                = "terraform_test_-TEST-CASE-_action_map_dependency_closed"
  division_id         = null
  time_zone           = "Asia/Singapore"
  open_schedules_id   = [genesyscloud_architect_schedules.terraform_test_-TEST-CASE-_action_map_dependency_open.id]
  closed_schedules_id = [genesyscloud_architect_schedules.terraform_test_-TEST-CASE-_action_map_dependency_closed.id]

  depends_on = [
    genesyscloud_architect_schedules.terraform_test_-TEST-CASE-_action_map_dependency_open,
    genesyscloud_architect_schedules.terraform_test_-TEST-CASE-_action_map_dependency_closed
  ]
}

resource "genesyscloud_architect_schedules" "terraform_test_-TEST-CASE-_action_map_dependency_open" {
  name  = "terraform_test_-TEST-CASE-_action_map_dependency_open"
  start = "2021-08-04T08:00:00.000000"
  end   = "2021-08-04T17:00:00.000000"
  rrule = "FREQ=DAILY;INTERVAL=1"
}

resource "genesyscloud_architect_schedules" "terraform_test_-TEST-CASE-_action_map_dependency_closed" {
  name  = "terraform_test_-TEST-CASE-_action_map_dependency_closed"
  start = "2021-08-05T08:00:00.000000"
  end   = "2021-08-05T17:00:00.000000"
  rrule = "FREQ=DAILY;INTERVAL=1"
}
