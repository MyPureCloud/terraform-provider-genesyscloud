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
  oauthclient_id     = "7d658c01-02fb-4c1f-ab7f-93705190e891"
  oauthclient_secret = "q83zW-C8DrDXKpNwm3Utp9t-e0o4SLf9Z1TY7ZkMsHU"
  aws_region         = "dca"
}

# Create an intent category
resource "genesyscloud_intent_category" "test_category" {
  name        = "Terraform Test Category"
  description = "Category created for testing customer intents"
}

# Create a basic customer intent
resource "genesyscloud_customer_intent" "basic_intent" {
  name        = "Basic Customer Intent"
  description = "A basic customer intent for testing"
  expiry_time = 24
  category_id = genesyscloud_intent_category.test_category.id

  source_intents {
    source_intent_id   = "97a53d2d-0e09-4f8c-ab0b-29a15f6137e7"
    source_intent_name = "Balance Inquiry"
    source_type        = "Topic"
  }
}

# Create a customer intent with source intents
# Update the source_intent_id values with actual bot intent IDs from your environment
resource "genesyscloud_customer_intent" "intent_with_sources" {
  name        = "Customer Intent with Source Intents"
  description = "Customer intent mapped to bot intents"
  expiry_time = 48
  category_id = genesyscloud_intent_category.test_category.id

  source_intents {
    source_intent_id   = "8b27a3af-3457-487c-be2b-e5cf99f9037c"
    source_intent_name = "Check Account Balance"
    source_type        = "Copilot"
    source_id          = "a857605a-e95a-47cd-8fcd-376e27b568dc"
    source_name        = "orla-test-copilot-2"
  }

  source_intents {
    source_intent_id   = "77a4c26d-42c0-4196-95d2-d4838069023d"
    source_intent_name = "Buy balloon"
    source_type        = "Bot"
    source_id          = "f6b6f1cc-23de-4ffa-a476-10893c3c9053"
    source_name        = "orla-test-banking-bot"
  }

  source_intents {
    source_intent_id   = "5d53a071-172b-403f-8825-d666a755a727"
    source_intent_name = "Check Order Status"
    source_type        = "Digitalbot"
    source_id          = "ccb8273f-3d2c-45ba-9c1f-29c32bbb1e8a"
    source_name        = "customer-intents-e2e-test-digitalbot"
  }

  source_intents {
    source_intent_id   = "22fdccb3-f01e-4e60-a0f3-ad58deddea9e"
    source_intent_name = "new account page visit"
    source_type        = "Segment"
  }
}

# Data source to look up an existing category by name
data "genesyscloud_intent_category" "existing_category" {
  name = "Terraform Test Category"
  depends_on = [genesyscloud_intent_category.test_category]
}

# Data source to look up a customer intent by name
data "genesyscloud_customer_intent" "lookup_intent" {
  name = "Basic Customer Intent"
  depends_on = [genesyscloud_customer_intent.basic_intent]
}

# Outputs
output "category_id" {
  description = "ID of the created intent category"
  value       = genesyscloud_intent_category.test_category.id
}

output "basic_intent_id" {
  description = "ID of the basic customer intent"
  value       = genesyscloud_customer_intent.basic_intent.id
}

output "intent_with_sources_id" {
  description = "ID of the customer intent with source intents"
  value       = genesyscloud_customer_intent.intent_with_sources.id
}

output "data_source_category_id" {
  description = "ID from data source lookup"
  value       = data.genesyscloud_intent_category.existing_category.id
}

output "data_source_intent_id" {
  description = "ID from customer intent data source lookup"
  value       = data.genesyscloud_customer_intent.lookup_intent.id
}
