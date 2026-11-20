data "genesyscloud_workforcemanagement_businessunits_activitycode" "example_activity_code" {
  business_unit_id = genesyscloud_workforcemanagement_businessunits.example_with_settings.id
  name             = "Example Activity Code"
}

# A default activity code has an empty name unless it was renamed, so it is selected by its category
data "genesyscloud_workforcemanagement_businessunits_activitycode" "example_default_activity_code" {
  business_unit_id = genesyscloud_workforcemanagement_businessunits.example_with_settings.id
  category         = "Break"
}
