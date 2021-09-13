terraform {
  required_version = "~> 0.14.0"
  required_providers {
    genesyscloud = {
      source  = "genesys.com/mypurecloud/genesyscloud"
      version = "~> 0.1.0"
    }
  }
}

provider "genesyscloud" {
  oauthclient_id = "df4cf7c9-bdcd-4c87-bb90-969455486dd1"
  oauthclient_secret = "1zjnIHkin-5UKH_u2dLbHsoax6K9kvj0ZNhi8wHJY6w"
  aws_region = "dca"
}

resource "genesyscloud_architect_flow" "test_flow1" {
  name   = "Terraform Flow Test-e2b9b3f6-5e84-4804-b381-d556864d764b"
  type = "INBOUNDCALL"
  filepath = "../resources/genesyscloud_architect_flow/inboundcall_flow_example.yaml"
  debug = false
  force_unlock = true
  recreate = true
}