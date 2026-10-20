resource "genesyscloud_webdeployments_deployment_identity_resolution" "example_deployment_identity_resolution" {
  deployment_id      = genesyscloud_webdeployments_deployment.example_deployment.id
  resolve_identities = true
  division_id        = data.genesyscloud_auth_division_home.home.id
  automerge_config {
    # Both flags are required whenever this block is present. Enabling
    # authenticated_web_messaging additionally requires authentication_settings with
    # enabled = true on the deployment's configuration, otherwise the platform silently
    # refuses it and the apply fails.
    authenticated_web_messaging = false
    web_tracking                = true
  }
}
