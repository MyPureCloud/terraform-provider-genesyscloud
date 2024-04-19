// Flow loglevel is still in beta and protected by a feature toggle.
// To enable flow loglevels in your org contact your Genesys Cloud account manager
resource "genesyscloud_flow_loglevel" "flowLogLevel" {
  flow_id        = genesyscloud_flow.flow.id
  flow_log_level = "Base"
}
