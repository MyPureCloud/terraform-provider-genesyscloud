resource "genesyscloud_agentic_virtual_agent_version_publish" "example_publish" {
  agent_id = genesyscloud_agentic_virtual_agent.example_agent.id
  version  = genesyscloud_agentic_virtual_agent_version.example_version.version
  status   = "ProductionReady"
}
