# Resolve the composite Terraform id for a stageplan (caseplan_id|stage_number|stageplan_id).
# Replace caseplan_id with your draft caseplan UUID or a reference such as genesyscloud_case_management_caseplan.main.id
data "genesyscloud_case_management_stageplan" "example_stageplan" {
  caseplan_id  = "00000000-0000-0000-0000-000000000000"
  stage_number = 1
}
