terraform {
  required_version = ">= 1.0.0"
  required_providers {
    genesyscloud = {
      source  = "genesys.com/mypurecloud/genesyscloud"
      version = "0.1.0"  # Use this for local sideload
    }
  }
}

provider "genesyscloud" {
  # Credentials will be read from environment variables
  # GENESYSCLOUD_OAUTHCLIENT_ID
  # GENESYSCLOUD_OAUTHCLIENT_SECRET
  # GENESYSCLOUD_REGION
}
