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

resource "genesyscloud_architect_flow" "supercool_flow" {
  filepath = "../resources/genesyscloud_architect_flow/inboundcall_flow_example.yaml"
}