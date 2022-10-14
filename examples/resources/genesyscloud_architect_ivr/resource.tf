resource "genesyscloud_architect_ivr" "sample_ivr" {
  name                  = "Sample IVR"
  description           = "A sample IVR configuration"
  dnis                  = ["+13175550000", "+13175550001"]
  open_hours_flow_id    = data.genesyscloud_flow.open-hours.id
  closed_hours_flow_id  = data.genesyscloud_flow.closed-hours.id
  holiday_hours_flow_id = data.genesyscloud_flow.holiday-hours.id
  schedule_group_id     = data.genesyscloud_architect_schedulegroups.group.id
}
