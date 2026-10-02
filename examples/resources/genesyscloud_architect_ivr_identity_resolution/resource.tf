resource "genesyscloud_architect_ivr_identity_resolution" "example_ivr_identity_resolution" {
  ivr_id             = genesyscloud_architect_ivr.sample_ivr.id
  resolve_identities = true
}