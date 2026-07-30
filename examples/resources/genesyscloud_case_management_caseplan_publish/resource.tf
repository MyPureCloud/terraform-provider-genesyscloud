# Replace caseplan_id with genesyscloud_case_management_caseplan.<name>.id (or import id).
# Use depends_on = [ ... stageplans, stepplans ... ] so publish runs after configuration PATCHes.
resource "genesyscloud_case_management_caseplan_publish" "example" {
  caseplan_id = "00000000-0000-0000-0000-000000000001"
}
