data "genesyscloud_organizations_me" "genesys_cloud_org" {}

output "genesys_cloud_org_id" {
  value = data.genesyscloud_organizations_me.genesys_cloud_org.id
}