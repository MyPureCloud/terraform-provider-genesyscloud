---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "genesyscloud_externalcontacts_organization Data Source - terraform-provider-genesyscloud"
subcategory: ""
description: |-
  Genesys Cloud external contacts organization data source. Select an external contacts organization by name
---

# genesyscloud_externalcontacts_organization (Data Source)

Genesys Cloud external contacts organization data source. Select an external contacts organization by name

## Example Usage

```terraform
data "genesyscloud_externalcontacts_organization" "organization" {
  name = "ABCNewsCompany"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) external contacts organization name

### Read-Only

- `id` (String) The ID of this resource.