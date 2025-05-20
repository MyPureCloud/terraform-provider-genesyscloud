resource "genesyscloud_outbound_callanalysisresponseset" "example_cars" {
  name                   = "Example Call Analysis Response Set"
  beep_detection_enabled = false
  responses {
    callable_person {
      name          = "Example Outbound Flow"
      data          = genesyscloud_flow.outbound_call_flow.id
      reaction_type = "transfer_flow"
    }
    callable_machine {
      reaction_type = "hangup"
    }
  }
}
