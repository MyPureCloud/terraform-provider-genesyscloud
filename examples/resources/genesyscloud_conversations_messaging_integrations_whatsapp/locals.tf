locals {
  dependencies = {
    resource = [
      "../genesyscloud_conversations_messaging_settings/resource.tf",
      "../genesyscloud_conversations_messaging_supportedcontent/resource.tf"
    ]
  }

  skip_if = {
    products_missing_any = ["messagingWhatsapp", "messagingPlatformWhatsApp"]
  }
}
