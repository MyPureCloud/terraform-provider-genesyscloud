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
