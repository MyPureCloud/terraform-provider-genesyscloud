# Replace caseplan_id with genesyscloud_case_management_caseplan.<name>.id
resource "genesyscloud_case_management_stageplan" "example" {
  caseplan_id  = "00000000-0000-0000-0000-000000000001"
  stage_number = 1
  name         = "Intake"
  description  = "First stage"
}
