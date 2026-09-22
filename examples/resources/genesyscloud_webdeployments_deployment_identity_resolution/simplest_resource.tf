resource "genesyscloud_webdeployments_deployment_identity_resolution" "example_deployment_identity_resolution" {
  deployment_id      = genesyscloud_webdeployments_deployment.example_deployment.id
  resolve_identities = false
}
