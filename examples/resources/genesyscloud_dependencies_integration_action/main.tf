terraform {
  # Version constraints are ignored when a dev override is active
  # required_version = ">= 1.0.0"
  required_providers {
    genesyscloud = {
      source  = "MyPureCloud/genesyscloud"
      # version = ">= 1.15.0"
    }
		random = {
			source = "hashicorp/random",
			version = ">= 3.7.2"
		}
		time = {
			source = "hashicorp/time",
			version = ">= 0.13.1",
		}
		tls = {
			source = "hashicorp/tls",
			version = "~> 4.0",
		}
  }
}

provider "genesyscloud" {
  # Set oAuth credentials via environment variables:
  # GENESYSCLOUD_OAUTHCLIENT_ID, GENESYSCLOUD_OAUTHCLIENT_SECRET, and GENESYSCLOUD_REGION
  # region = "us-east-1"
}
