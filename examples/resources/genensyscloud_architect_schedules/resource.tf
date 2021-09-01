resource "genesyscloud_architect_schedules" "test_schedule" {
  name        = "CX Code Schedule"
  description = "Sample Scedhule by CX as Code"
  start       = "2021-08-04T08:00:00.000000"
  end         = "2021-08-04T17:00:00.000000"
  rrule       = "FREQ=DAILY;INTERVAL=1"
}