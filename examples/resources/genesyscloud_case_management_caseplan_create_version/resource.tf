# Replace caseplan_id with genesyscloud_case_management_caseplan.<name>.id.
# Run create_version after a publish when there is no open draft (see resource description).
resource "genesyscloud_case_management_caseplan_publish" "example" {
  caseplan_id = "00000000-0000-0000-0000-000000000001"
}

resource "genesyscloud_case_management_caseplan_create_version" "example" {
  caseplan_id = "00000000-0000-0000-0000-000000000001"
  depends_on  = [genesyscloud_case_management_caseplan_publish.example]
}
