# Replace caseplan_id and worktype_id with real resource references.
resource "genesyscloud_case_management_stepplan" "example" {
  caseplan_id   = "00000000-0000-0000-0000-000000000001"
  stage_number  = 1
  name          = "Step 1"
  description   = "Workitem step"
  activity_type = "workitem"
  workitem_settings {
    worktype_id = "00000000-0000-0000-0000-000000000002"
  }
}
