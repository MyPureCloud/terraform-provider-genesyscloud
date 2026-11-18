# Validate a flow by ID and fail the plan if it has errors
data "genesyscloud_flow_validation" "inbound" {
  flow_id       = genesyscloud_flow.inbound_call_flow.id
  fail_on_error = true
}

# Validate a specific version of a flow selected by name and type, and report the results without failing
data "genesyscloud_flow_validation" "inqueue" {
  name    = "Default In-Queue Flow"
  type    = "inqueuecall"
  version = "1.0"
}

output "inqueue_flow_valid" {
  value = data.genesyscloud_flow_validation.inqueue.valid
}

output "inqueue_flow_errors" {
  value = data.genesyscloud_flow_validation.inqueue.errors
}
