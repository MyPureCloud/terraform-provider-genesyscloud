data "genesyscloud_auth_division_home" "home" {}

resource "genesyscloud_intent_category" "example_caseplan_category" {
  name        = "Example caseplan intent category"
  description = "Category for caseplan documentation example"
}

resource "genesyscloud_customer_intent" "example_caseplan_intent" {
  name        = "Example caseplan customer intent"
  description = "Customer intent for caseplan example"
  expiry_time = 24
  category_id = genesyscloud_intent_category.example_caseplan_category.id
}

resource "genesyscloud_task_management_workitem_schema" "caseplan_example_schema" {
  name        = "caseplan_example_schema"
  description = "Task management schema bound to caseplan data_schemas"
  enabled     = true
  properties = jsonencode({
    note_text = {
      allOf       = [{ "$ref" = "#/definitions/text" }]
      title       = "Note"
      description = "Example text"
      minLength   = 0
      maxLength   = 100
    }
  })
}

resource "genesyscloud_user" "example_caseplan_owner" {
  email       = "caseplan_doc_example_owner@example.com"
  name        = "Example caseplan default owner"
  password    = "TerraformDocExample1!"
  division_id = data.genesyscloud_auth_division_home.home.id
}

resource "genesyscloud_case_management_caseplan" "example" {
  name                            = "Example caseplan"
  description                     = "Example case management caseplan"
  division_id                     = data.genesyscloud_auth_division_home.home.id
  reference_prefix                = "EXPL"
  default_due_duration_in_seconds = 1296000
  default_ttl_seconds             = 31536000

  customer_intent {
    id = genesyscloud_customer_intent.example_caseplan_intent.id
  }

  default_case_owner {
    id = genesyscloud_user.example_caseplan_owner.id
  }

  data_schema {
    id      = genesyscloud_task_management_workitem_schema.caseplan_example_schema.id
    version = floor(genesyscloud_task_management_workitem_schema.caseplan_example_schema.version)
  }
}
