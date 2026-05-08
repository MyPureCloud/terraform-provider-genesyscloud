# Resolve the composite Terraform id for a stepplan (caseplan_id|stage_number|stepplan_id).
# Replace caseplan_id with your draft caseplan UUID or genesyscloud_case_management_caseplan.<name>.id
data "genesyscloud_case_management_stepplan" "example_stepplan" {
  caseplan_id  = "00000000-0000-0000-0000-000000000000"
  stage_number = 1
}
