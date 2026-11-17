resource "genesyscloud_authorization_roles_settings" "settings" {
  backfill_enabled                   = true
  authorization_grant_division_aware = false

  # Genesys Cloud enables this by default. Setting it to false stops paired Genesys Services
  # organizations and Genesys CIDR ranges from bypassing attribute-based access control
  # policies, which can block Genesys support access to this organization.
  genesys_org_policy_bypass = true
}
