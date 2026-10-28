locals {
  dependencies = {
    resource = [
      "../genesyscloud_conversations_messaging_integrations_apple/resource.tf",
      "../../data-sources/genesyscloud_auth_division_home/data-source.tf",
      "../genesyscloud_conversations_messaging_settings/resource.tf",
      "../genesyscloud_conversations_messaging_supportedcontent/resource.tf"
    ]
  }
}