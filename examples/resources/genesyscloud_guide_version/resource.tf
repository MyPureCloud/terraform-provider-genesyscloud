resource "genesyscloud_guide_version" "sample-guide" {
  guide_id    = "data.genesyscloud_guide.id"
  instruction = "This is a test Instruction"
  variables {
    name        = "TestVariable"
    type        = "String"
    scope       = "InputAndOutput"
    description = "This is a test Description"
  }
  resources {
    data_action {
      data_action_id = "DataActionId"
      label          = "Genesys Cloud Data Actions (1)"
      description    = "This is a test Description"
    }
    data_action {
      data_action_id = "DataActionId"
      label          = "Genesys Cloud Data Actions (1)"
      description    = "This is a test Description"
    }
  }
  state = "ProductionReady"
}