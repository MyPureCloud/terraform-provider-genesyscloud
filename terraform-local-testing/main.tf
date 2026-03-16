terraform {
  required_version = ">= 1.0.0"
  required_providers {
    genesyscloud = {
      source  = "mypurecloud/genesyscloud"
      version = ">= 1.6.0"
    }
  }
}

provider "genesyscloud" {
  oauthclient_id     = "d68108fa-cf2c-4976-8f57-c09b61fdc75f"
  oauthclient_secret = "iySpgjO1HpOjjl7hWpjTcVLsBNQOdQ3GoLsoXZUx9CY"
  aws_region         = "us-west-2"  # Change to your region (e.g., us-east-1, eu-west-1, ap-southeast-2)
}

# Add your Genesys Cloud resources here
# For example:
# resource "genesyscloud_user" "example" {
#   email = "user@example.com"
#   name  = "John Doe"
# }