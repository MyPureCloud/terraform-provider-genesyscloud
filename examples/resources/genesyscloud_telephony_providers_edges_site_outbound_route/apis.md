- [GET /api/v2/telephony/providers/edges/sites](https://developer.genesys.cloud/api/rest/v2/telephonyprovidersedge/#get-api-v2-telephony-providers-edges-sites)
- [GET /api/v2/telephony/providers/edges/sites/{siteId}/outboundroutes](https://developer.genesys.cloud/api/rest/v2/telephonyprovidersedge/#get-api-v2-telephony-providers-edges-sites--siteId--outboundroutes)
- [POST /api/v2/telephony/providers/edges/sites/{siteId}/outboundroutes](https://developer.genesys.cloud/api/rest/v2/telephonyprovidersedge/#post-api-v2-telephony-providers-edges-sites--siteId--outboundroutes)
- [DELETE /api/v2/telephony/providers/edges/sites/{siteId}/outboundroutes/{outboundRouteId}](https://developer.genesys.cloud/api/rest/v2/telephonyprovidersedge/#delete-api-v2-telephony-providers-edges-sites--siteId--outboundroutes--outboundRouteId-)
- [PUT /api/v2/telephony/providers/edges/sites/{siteId}/outboundroutes/{outboundRouteId}](https://developer.genesys.cloud/api/rest/v2/telephonyprovidersedge/#put-api-v2-telephony-providers-edges-sites--siteId--outboundroutes--outboundRouteId-)

#### Compatibility Note

In versions 1.39.0 to 1.48.0 of the provider, this resource was constructed with a different structure. The current version introduces structural changes that are not backwards compatible with those earlier versions.

These changes are currently controlled by a feature flag and are not yet the default behavior. This allows for a phased implementation and thorough testing of this resource before full release.

If you're upgrading from an earlier version, please be aware of these structural changes and consult these examples on how to migrate your configuration.

## Export Behavior

When exporting this resource, please be aware of the following behavior:

If the associated Genesys Cloud Telephony Site is configured as a `managed` resource:

- This resource will be exported as a data object
- Updates and modifications to this resource and its child dependencies will not be allowed through through this provider or the Genesys Cloud API
- This limitation is enforced by the Genesys Cloud API itself

This behavior ensures consistency with Genesys Cloud's management policies for managed telephony sites.
