resource "genesyscloud_workforcemanagement_businessunits_activitycode" "team_meeting" {
  business_unit_id        = "00000000-0000-0000-0000-000000000000"
  name                    = "Team Meeting"
  category                = "Meeting"
  length_in_minutes       = 30
  counts_as_paid_time     = true
  counts_as_work_time     = false
  counts_toward_shrinkage = true
  planned_shrinkage       = true
  interruptible           = false
}
