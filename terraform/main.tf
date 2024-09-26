

terraform {
  required_providers {
    genesyscloud = {
      source  = "registry.terraform.io/mypurecloud/genesyscloud"
      version = "1.40.0"
    }
  }
}

provider "genesyscloud" {
		oauthclient_id = "5d0c120e-f003-4ff9-a6a6-03b41c874f59"
		oauthclient_secret = "k5d5UexCHK596Ptzjt22bc6ERH0JOfgcp_ExKe7zGso"
		aws_region = "eu-west-1"
		sdk_debug          = true
		sdk_debug_format   = "Json"
	}
	resource "genesyscloud_tf_export" "export" {
		directory          = "./genesyscloud-will"
		include_filter_resources     = ["genesyscloud_routing_wrapupcode"]


	include_state_file = true
	enable_dependency_resolution = true
	split_files_by_resource = true
	export_as_hcl = true
	}