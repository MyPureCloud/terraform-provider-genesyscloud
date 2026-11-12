# Terraform manages the published caseplan. Every apply that changes a versioned attribute
# (due/ttl, default_case_owner, intake_settings, stageplan) creates or reuses a draft, writes the
# differences, and publishes. name and description are patched without publishing.
# division_id, customer_intent, reference_prefix and data_schema cannot change after the first publish.
resource "genesyscloud_case_management_caseplan" "example" {
  name                            = "Example caseplan"
  description                     = "Example case management caseplan"
  division_id                     = data.genesyscloud_auth_division_home.home.id
  reference_prefix                = "EXPL"
  default_due_duration_in_seconds = 1296000
  default_ttl_seconds             = 31536000

  customer_intent {
    id = genesyscloud_intents_customerintents.example_customer_intent.id
  }

  default_case_owner {
    id = genesyscloud_user.example_user.id
  }

  data_schema {
    id = genesyscloud_task_management_workitem_schema.example_schema.id
  }

  intake_settings {
    property      = "custom_attribute_1_text"
    required      = false
    display_order = 1
  }

  # 1-5 stageplans, in order. Each stageplan has exactly one stepplan.
  stageplan {
    name        = "Intake"
    description = "Capture the request"
    stepplan {
      name = "Triage"
    }
  }

  stageplan {
    name = "Investigate"
    stepplan {
      name          = "Investigate request"
      activity_type = "Workitem"
      workitem_settings {
        worktype_id = genesyscloud_task_management_worktype.example_worktype.id
      }
    }
  }

  stageplan {
    name = "Resolve"
    stepplan {
      name = "Close out"
    }
  }
}
