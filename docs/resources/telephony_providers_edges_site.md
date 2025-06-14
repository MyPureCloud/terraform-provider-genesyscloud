---
page_title: "genesyscloud_telephony_providers_edges_site Resource - terraform-provider-genesyscloud"
subcategory: ""
description: |-
  Genesys Cloud Site
---
# genesyscloud_telephony_providers_edges_site (Resource)

Genesys Cloud Site

## API Usage
The following Genesys Cloud APIs are used by this resource. Ensure your OAuth Client has been granted the necessary scopes and permissions to perform these operations:

- [GET /api/v2/telephony/providers/edges/sites](https://developer.genesys.cloud/api/rest/v2/telephonyprovidersedge/#get-api-v2-telephony-providers-edges-sites)
- [POST /api/v2/telephony/providers/edges/sites](https://developer.genesys.cloud/api/rest/v2/telephonyprovidersedge/#post-api-v2-telephony-providers-edges-sites)
- [DELETE /api/v2/telephony/providers/edges/sites/{siteId}](https://developer.genesys.cloud/api/rest/v2/telephonyprovidersedge/#delete-api-v2-telephony-providers-edges-sites--siteId-)
- [GET /api/v2/telephony/providers/edges/sites/{siteId}](https://developer.genesys.cloud/api/rest/v2/telephonyprovidersedge/#get-api-v2-telephony-providers-edges-sites--siteId-)
- [PUT /api/v2/telephony/providers/edges/sites/{siteId}](https://developer.genesys.cloud/api/rest/v2/telephonyprovidersedge/#put-api-v2-telephony-providers-edges-sites--siteId-)
- [GET /api/v2/telephony/providers/edges/sites/{siteId}/numberplans](https://developer.genesys.cloud/api/rest/v2/telephonyprovidersedge/#get-api-v2-telephony-providers-edges-sites--siteId--numberplans)
- [PUT /api/v2/telephony/providers/edges/sites/{siteId}/numberplans](https://developer.genesys.cloud/api/rest/v2/telephonyprovidersedge/#put-api-v2-telephony-providers-edges-sites--siteId--numberplans)
- [GET /api/v2/telephony/providers/edges/sites/{siteId}/outboundroutes](https://developer.genesys.cloud/api/rest/v2/telephonyprovidersedge/#get-api-v2-telephony-providers-edges-sites--siteId--outboundroutes)
- [POST /api/v2/telephony/providers/edges/sites/{siteId}/outboundroutes](https://developer.genesys.cloud/api/rest/v2/telephonyprovidersedge/#post-api-v2-telephony-providers-edges-sites--siteId--outboundroutes)
- [DELETE /api/v2/telephony/providers/edges/sites/{siteId}/outboundroutes/{outboundRouteId}](https://developer.genesys.cloud/api/rest/v2/telephonyprovidersedge/#delete-api-v2-telephony-providers-edges-sites--siteId--outboundroutes--outboundRouteId-)
- [PUT /api/v2/telephony/providers/edges/sites/{siteId}/outboundroutes/{outboundRouteId}](https://developer.genesys.cloud/api/rest/v2/telephonyprovidersedge/#put-api-v2-telephony-providers-edges-sites--siteId--outboundroutes--outboundRouteId-)

## Export Behavior

When exporting this resource, please be aware of the following behavior:

If the Genesys Cloud Telephony Site is configured as a `managed` resource:

- This resource will be exported as a data resource
- Updates and modifications to this resource and its child dependencies will not be allowed through through this provider or the Genesys Cloud API
- This limitation is enforced by the Genesys Cloud API itself

This behavior ensures consistency with Genesys Cloud's management policies for managed telephony sites.

## Breaking Changes in 1.61.0

### Removal of `outbound_routes` attribute from `genesyscloud_telephony_providers_edges_site`

In version 1.39.0 we introduced the `genesyscloud_telephony_providers_edges_site_outbound_route` resource to replace the `outbound_routes` attributes in the `genesyscloud_telephony_providers_edges_site` resource. This was held behind a feature toggle for a number of releases, with the attribute marked with a deprecated notice.

As of version 1.61.0, the `outbound_routes` attribute has been completely removed from the `genesyscloud_telephony_providers_edges_site` resource. This functionality is moved entirely to the dedicated resource: `genesyscloud_telephony_providers_edges_site_outbound_route`. A migration is required for any existing `outbound_routes` attributes.

#### Migration Steps

When upgrading to version 1.61.0, the provider will automatically migrate your state by removing the `outbound_routes` configuration from the `genesyscloud_telephony_providers_edges_site` resource. However, you will need to manually:

1. Run `terraform init -upgrade` to get the latest provider version
2. The provider will output the necessary configuration blocks and import commands for your existing outbound routes
3. Add the configuration blocks to your Terraform configuration
4. Run the provided import commands
5. Run `terraform plan` to verify the changes

#### Example Migration

##### Prior to version 1.61.0

```hcl
resource "genesyscloud_telephony_providers_edges_site" "example" {
  name        = "My Site"
  description = "My test site"
  location_id = genesyscloud_location.location.id
  media_model = "Cloud"

  outbound_routes {
    name                   = "Test outbound route"
    description           = "Test outbound route description"
    classification_types  = ["International"]
    enabled              = true
    distribution         = "SEQUENTIAL"
    external_trunk_base_ids = [
      genesyscloud_telephony_providers_edges_trunkbasesettings.trunk-base.id
    ]
  }
}
```

##### Version 1.61.0 and on

```hcl
resource "genesyscloud_telephony_providers_edges_site" "example" {
  name = "My Site"
  description = "My test site"
  location_id = genesyscloud_location.location.id
  media_model = "Cloud"
}

resource "genesyscloud_telephony_providers_edges_site_outbound_route" "example" {
  site_id = genesyscloud_telephony_providers_edges_site.example.id
  name = "Test outbound route"
  description = "Test outbound route description"
  classification_types = ["International"]
  enabled = true
  distribution = "SEQUENTIAL"
  external_trunk_base_ids = [
    genesyscloud_telephony_providers_edges_trunkbasesettings.trunk-base.id
  ]
}
```

The example above shows how to migrate from the old configuration to the new configuration. The state will be automatically migrated when you upgrade to version 1.61.0, but you will need to update your Terraform configuration files manually.


## Example Usage

```terraform
resource "genesyscloud_telephony_providers_edges_site" "site" {
  name                            = "example site"
  description                     = "example site description"
  location_id                     = genesyscloud_location.hq.id
  media_model                     = "Cloud"
  media_regions_use_latency_based = true
  edge_auto_update_config {
    time_zone = "America/New_York"
    rrule     = "FREQ=WEEKLY;BYDAY=SU"
    start     = "2021-08-08T08:00:00.000000"
    end       = "2021-08-08T11:00:00.000000"
  }
  number_plans {
    name           = "numberList plan"
    classification = "numberList classification"
    match_type     = "numberList"
    numbers {
      start = "114"
      end   = "115"
    }
  }
  number_plans {
    name           = "digitLength plan"
    classification = "digitLength classification"
    match_type     = "digitLength"
    digit_length {
      start = "6"
      end   = "8"
    }
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `location_id` (String) Site location ID
- `media_model` (String) Media model for the site Valid Values: Premises, Cloud. Changing the media_model attribute will cause the site object to be dropped and created with a new ID.
- `name` (String) The name of the entity.

### Optional

- `caller_id` (String) The caller ID value for the site. The callerID must be a valid E.164 formatted phone number
- `caller_name` (String) The caller name for the site
- `description` (String) The resource's description.
- `edge_auto_update_config` (Block List, Max: 1) Recurrence rule, time zone, and start/end settings for automatic edge updates for this site (see [below for nested schema](#nestedblock--edge_auto_update_config))
- `media_regions` (List of String) The ordered list of AWS regions through which media can stream. A full list of available media regions can be found at the GET /api/v2/telephony/mediaregions endpoint
- `media_regions_use_latency_based` (Boolean) Latency based on media region Defaults to `false`.
- `number_plans` (Block List) Number plans for the site. The order of the plans in the resource file determines the priority of the plans. Specifying number plans will not result in the default plans being overwritten. (see [below for nested schema](#nestedblock--number_plans))
- `primary_sites` (List of String) Used for primary phone edge assignment on physical edges only.  List of primary sites the phones can be assigned to. If no primary_sites are defined, the site id for this site will be used as the primary site id.
- `secondary_sites` (List of String) Used for secondary phone edge assignment on physical edges only.  List of secondary sites the phones can be assigned to.  If no primary_sites or secondary_sites are defined then the current site will defined as primary and secondary.
- `set_as_default_site` (Boolean) Set this site as the default site for the organization. Only one genesyscloud_telephony_providers_edges_site resource should be set as the default. Defaults to `false`.

### Read-Only

- `id` (String) The ID of this resource.
- `managed` (Boolean) Is this site managed by Genesys Cloud

<a id="nestedblock--edge_auto_update_config"></a>
### Nested Schema for `edge_auto_update_config`

Required:

- `end` (String) Date time is represented as an ISO-8601 string without a timezone. For example: yyyy-MM-ddTHH:mm:ss.SSS
- `rrule` (String) A reoccurring rule for updating the Edges assigned to the site. The only supported frequencies are daily and weekly. Weekly frequencies require a day list with at least oneday specified. All other configurations are not supported.
- `start` (String) Date time is represented as an ISO-8601 string without a timezone. For example: yyyy-MM-ddTHH:mm:ss.SSS
- `time_zone` (String) The timezone of the window in which any updates to the edges assigned to the site can be applied. The minimum size of the window is 2 hours.


<a id="nestedblock--number_plans"></a>
### Nested Schema for `number_plans`

Required:

- `classification` (String) Used to classify this number plan
- `match_type` (String)
- `name` (String) The name of the entity.

Optional:

- `digit_length` (Block List, Max: 1) Allowed values are between 1-20 digits. (see [below for nested schema](#nestedblock--number_plans--digit_length))
- `match_format` (String) Use regular expression capture groups to build the normalized number
- `normalized_format` (String) Use regular expression capture groups to build the normalized number
- `numbers` (Block List) Numbers must be 2-9 digits long. Numbers within ranges must be the same length. (e.g. 888, 888-999, 55555-77777, 800). (see [below for nested schema](#nestedblock--number_plans--numbers))

<a id="nestedblock--number_plans--digit_length"></a>
### Nested Schema for `number_plans.digit_length`

Optional:

- `end` (String)
- `start` (String)


<a id="nestedblock--number_plans--numbers"></a>
### Nested Schema for `number_plans.numbers`

Optional:

- `end` (String)
- `start` (String)

