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

resource "genesyscloud_tf_export" "export_by_regex" {
  directory                      = "./export-regex"
  include_state_file             = true
  include_filter_resources       = ["genesyscloud_flow::Email Decryption Flow"]
  export_format                  = "json"
  split_files_by_resource        = true
  enable_dependency_resolution   = true
}

output "export_directory" {
  value = genesyscloud_tf_export.export_by_regex.directory
}
