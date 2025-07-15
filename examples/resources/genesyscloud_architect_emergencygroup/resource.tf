resource "genesyscloud_architect_emergencygroup" "emergency-group" {
  name        = "CX as Code Emergency Group"
  description = "Sample Emergency Group by CX as Code"
  emergency_call_flows {
    emergency_flow_id = genesyscloud_flow.inbound_call_flow.id
    ivr_ids           = [genesyscloud_architect_ivr.sample_ivr.id]
  }
}
