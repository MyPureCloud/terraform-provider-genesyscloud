resource "genesyscloud_architect_schedulegroups" "sample_schedule_groups" {
  name                 = "CX as Code Schedule Group"
  description          = "Sample Schedule Group by CX as Code"
  time_zone            = "Asia/Singapore"
  open_schedules_id    = [genesyscloud_architect_schedules.open1.id, genesyscloud_architect_schedules.open2.id]
  closed_schedules_id  = [genesyscloud_architect_schedules.closed.id]
  holiday_schedules_id = [genesyscloud_architect_schedules.holiday.id]
}