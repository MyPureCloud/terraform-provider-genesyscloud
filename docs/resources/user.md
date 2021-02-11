---
page_title: "genesyscloud_user Resource - terraform-provider-genesyscloud"
subcategory: ""
description: |-
  Genesys Cloud User
---

# Resource `genesyscloud_user`

Genesys Cloud User

## Example Usage

```terraform
resource "genesyscloud_user" "test_user" {
  email       = "test@example.com"
  name        = "Test User"
  password    = "initial-password"
  division_id = "505e1036-6f04-405c-a630-de94a8ad2eb8"
}
```

## Schema

### Required

- **email** (String) User's email and username.
- **name** (String) User's full name.

### Optional

- **division_id** (String) The division to which this user will belong. If not set, the home division will be used.
- **id** (String) The ID of this resource.
- **password** (String, Sensitive) User's password. If specified, this is only set on user create.


