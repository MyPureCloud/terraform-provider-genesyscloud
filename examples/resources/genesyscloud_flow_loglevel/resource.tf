resource "genesyscloud_flow_loglevel" "flowLogLevel" {
  flow_id        = genesyscloud_flow.flow.id
  flow_log_level = "Base"
}
