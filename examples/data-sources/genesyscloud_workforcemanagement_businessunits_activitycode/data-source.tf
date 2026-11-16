data "genesyscloud_workforcemanagement_businessunits_activitycode" "example_activity_code" {
  business_unit_id = genesyscloud_workforcemanagement_businessunits.example_with_settings.id
  name             = "Example Activity Code"
}
