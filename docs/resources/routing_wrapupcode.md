---
page_title: "genesyscloud_routing_wrapupcode Resource - terraform-provider-genesyscloud"
subcategory: ""
description: |-
  Genesys Cloud Routing Wrapup Code
---
# genesyscloud_routing_wrapupcode (Resource)

Genesys Cloud Routing Wrapup Code

## API Usage
The following Genesys Cloud APIs are used by this resource. Ensure your OAuth Client has been granted the necessary scopes and permissions to perform these operations:

* [GET /api/v2/routing/wrapupcodes](https://developer.mypurecloud.com/api/rest/v2/routing/#get-api-v2-routing-wrapupcodes)
* [GET /api/v2/routing/wrapupcodes/{codeId}](https://developer.mypurecloud.com/api/rest/v2/routing/#get-api-v2-routing-wrapupcodes--codeId-)
* [POST /api/v2/routing/wrapupcodes](https://developer.mypurecloud.com/api/rest/v2/routing/#post-api-v2-routing-wrapupcodes)
* [PUT /api/v2/routing/wrapupcodes/{codeId}](https://developer.mypurecloud.com/api/rest/v2/routing/#put-api-v2-routing-wrapupcodes--codeId-)
* [DELETE /api/v2/routing/wrapupcodes/{codeId}](https://developer.mypurecloud.com/api/rest/v2/routing/#delete-api-v2-routing-wrapupcodes--codeId-)

## Example Usage

```terraform
resource "genesyscloud_routing_wrapupcode" "win" {
  name        = "Win"
  description = "Win test description"
}
resource "genesyscloud_routing_wrapupcode" "unknown" {
  name        = "Unknown"
  description = "Unknown test description"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Wrapup Code name.

### Optional

- `description` (String) The wrap-up code description.
- `division_id` (String) The division to which this routing wrapupcode will belong. If not set, * will be used to indicate all divisions.

### Read-Only

- `id` (String) The ID of this resource.

