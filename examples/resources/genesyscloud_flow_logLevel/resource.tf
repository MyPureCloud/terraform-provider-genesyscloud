resource "genesyscloud_flow_loglevel" "flowLogLevel" {
  flow_id        = "UUID"
  flow_log_level = "Base"
  flow_characteristics {
    execution_items         = "true"
    execution_input_outputs = "false"
    communications          = "false"
    event_error             = "true"
    event_warning           = "true"
    event_other             = "false"
    variables               = "false"
    names                   = "false"
  }
}
