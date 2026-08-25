resource "genesyscloud_access_policy" "example_deny_role_grants" {
  name            = "Cannot Grant New Roles"
  description     = "Prevents users from granting new roles to other users"
  target_resource = "authorization:grant:add"
  effect          = "DENY"
  enabled         = true
  subject_type    = "ALL"

  condition_json = jsonencode({
    "and" : [
      {
        "attribute" : "subject.role.name",
        "operator" : "eq",
        "value" : "employee"
      }
    ]
  })
}
