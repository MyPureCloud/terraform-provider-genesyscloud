---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "genesyscloud_conversations_messaging_integrations_whatsapp Data Source - terraform-provider-genesyscloud"
subcategory: ""
description: |-
  Genesys Cloud conversations messaging integrations whatsapp data source. Select an conversations messaging integrations whatsapp by name
---

# genesyscloud_conversations_messaging_integrations_whatsapp (Data Source)

Genesys Cloud conversations messaging integrations whatsapp data source. Select an conversations messaging integrations whatsapp by name

## Example Usage

```terraform
data "genesyscloud_conversations_messaging_integrations_whatsapp" "integration_whatsapp" {
  name = "integration_whatsapp data"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) conversations messaging integrations whatsapp name

### Read-Only

- `id` (String) The ID of this resource.
