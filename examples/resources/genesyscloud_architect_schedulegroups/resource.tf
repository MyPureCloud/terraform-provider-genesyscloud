resource "genesyscloud_architect_schedulegroups" "test_schedule_groups" {
  name                 = "CX as Code Schedule Group"
  description          = "Sample Schedule Group by CX as Code"
  time_zone            = "Asia/Singapore"
  open_schedules_id    = ["d76457c0-3331-4e43-a96c-24e7bbc9a4ee", "84b8d87b-fdc0-4e01-97b1-482578866d2f"]
  closed_schedules_id  = ["f806b9a1-698f-4641-b0a8-289d1fdfd2eb", "e7bbdbf6-686d-4363-aca4-fb0e913667d3"]
  holiday_schedules_id = ["5b637b61-a6da-43b8-b739-9cab4b31c02a"]
}   