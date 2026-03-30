terraform {
  required_providers {
    genesyscloud = {
      source = "genesys.com/mypurecloud/genesyscloud"
    }
  }
}

provider "genesyscloud" {
  # Credentials will be read from environment variables:
  # GENESYSCLOUD_OAUTHCLIENT_ID
  # GENESYSCLOUD_OAUTHCLIENT_SECRET
  # GENESYSCLOUD_REGION
}

resource "genesyscloud_tf_export" "export_by_id" {
  directory                      = "./export-by-id"
  include_state_file             = true
  include_filter_resources_by_id = ["genesyscloud_flow::b84cbae3-7c54-45dc-ade0-7a30fbccf996"]
  export_format                  = "json"
  split_files_by_resource        = true
  enable_dependency_resolution   = true
}

output "export_directory" {
  value = genesyscloud_tf_export.export_by_id.directory
}
