resource "genesyscloud_workforcemanagement_businessunits_activitycode" "example_activity_code" {
  business_unit_id        = genesyscloud_workforcemanagement_businessunits.example_with_settings.id
  name                    = "Example Activity Code"
  category                = "Meeting"
  length_in_minutes       = 30
  counts_as_paid_time     = true
  counts_as_work_time     = false
  counts_toward_shrinkage = true
  planned_shrinkage       = true
  interruptible           = false
}
