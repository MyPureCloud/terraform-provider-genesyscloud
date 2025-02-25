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
