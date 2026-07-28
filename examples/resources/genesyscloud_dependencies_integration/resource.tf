resource "genesyscloud_integration" "hackathon_2026_integration" {
  intended_state   = "ENABLED"
  integration_type = "purecloud-data-actions"
  config {
    name      = "Demo Integration Dependencies - Hackathon 2026"
  }
}
