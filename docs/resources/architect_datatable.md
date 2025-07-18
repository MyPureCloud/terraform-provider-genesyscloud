---
page_title: "genesyscloud_architect_datatable Resource - terraform-provider-genesyscloud"
subcategory: ""
description: |-
  Genesys Cloud Architect Datatables
---
# genesyscloud_architect_datatable (Resource)

Genesys Cloud Architect Datatables

## API Usage
The following Genesys Cloud APIs are used by this resource. Ensure your OAuth Client has been granted the necessary scopes and permissions to perform these operations:

* [GET /api/v2/flows/datatables](https://developer.mypurecloud.com/api/rest/v2/architect/#get-api-v2-flows-datatables)
* [POST /api/v2/flows/datatables](https://developer.mypurecloud.com/api/rest/v2/architect/#post-api-v2-flows-datatables)
* [GET /api/v2/flows/datatables/{datatableId}](https://developer.mypurecloud.com/api/rest/v2/architect/#get-api-v2-flows-datatables--datatableId-)
* [PUT /api/v2/flows/datatables/{datatableId}](https://developer.mypurecloud.com/api/rest/v2/architect/#put-api-v2-flows-datatables--datatableId-)
* [DELETE /api/v2/flows/datatables/{datatableId}](https://developer.mypurecloud.com/api/rest/v2/architect/#delete-api-v2-flows-datatables--datatableId-)

## Example Usage

```terraform
resource "genesyscloud_architect_datatable" "customers" {
  name        = "Customers"
  division_id = data.genesyscloud_auth_division_home.home.id
  description = "Table of Customers"
  properties {
    name  = "key"
    type  = "string"
    title = "Email"
  }
  properties {
    name  = "identifier"
    type  = "integer"
    title = "Customer Identifier"
  }
  properties {
    name    = "deleted"
    type    = "boolean"
    title   = "Is Deleted"
    default = "false"
  }
  properties {
    name  = "address"
    type  = "string"
    title = "Address"
  }
  properties {
    name  = "vip"
    type  = "boolean"
    title = "VIP"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Name of the architect_datatable.
- `properties` (Block List, Min: 1) Schema properties of the architect_datatable. This must at a minimum contain a string property 'key' that will serve as the row key. Properties cannot be removed from a schema once they have been added (see [below for nested schema](#nestedblock--properties))

### Optional

- `description` (String) Description of the architect_datatable.
- `division_id` (String) The division to which this architect_datatable will belong. If not set, the home division will be used.

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--properties"></a>
### Nested Schema for `properties`

Required:

- `name` (String) Name of the property.
- `type` (String) Type of the property (boolean | string | integer | number).

Optional:

- `default` (String) Default value of the property. This is converted to the proper type for non-strings (e.g. set 'true' or 'false' for booleans).
- `title` (String) Display title of the property.

