# Replace data_schema id with a genesyscloud_task_management_workitem_schema id; set version to that schema's version.
# Optional: division_id, customer_intent, default_case_owner, intake_settings (see resource documentation).
resource "genesyscloud_case_management_caseplan" "example" {
  name                            = "Example caseplan"
  description                     = "Example case management caseplan"
  reference_prefix                = "EXPL"
  default_due_duration_in_seconds = 1296000
  default_ttl_seconds             = 31536000

  data_schema {
    id      = "00000000-0000-0000-0000-000000000001"
    version = 1
  }
}
