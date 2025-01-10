- [GET /api/v2/telephony/providers/edges/trunkbasesettings](https://developer.genesys.cloud/api/rest/v2/telephonyprovidersedge/#get-api-v2-telephony-providers-edges-trunkbasesettings)
- [POST /api/v2/telephony/providers/edges/trunkbasesettings](https://developer.genesys.cloud/api/rest/v2/telephonyprovidersedge/#post-api-v2-telephony-providers-edges-trunkbasesettings)
- [GET /api/v2/telephony/providers/edges/trunkbasesettings/{trunkBaseSettingsId}](https://developer.genesys.cloud/api/rest/v2/telephonyprovidersedge/#get-api-v2-telephony-providers-edges-trunkbasesettings--trunkBaseSettingsId-)
- [DELETE /api/v2/telephony/providers/edges/trunkbasesettings/{trunkBaseSettingsId}](https://developer.genesys.cloud/api/rest/v2/telephonyprovidersedge/#delete-api-v2-telephony-providers-edges-trunkbasesettings--trunkBaseSettingsId-)
- [PUT /api/v2/telephony/providers/edges/trunkbasesettings/{trunkBaseSettingsId}](https://developer.genesys.cloud/api/rest/v2/telephonyprovidersedge/#put-api-v2-telephony-providers-edges-trunkbasesettings--trunkBaseSettingsId-)

## Export Behavior

### Managed Trunk Base Settings

The following Trunk Base Settings are managed directly by Genesys Cloud and will be exported as data objects:

- Cloud Proxy Tie TrunkBase for EdgeGroup
- Direct Tie TrunkBase for EdgeGroup
- Genesys Cloud - CDM SIP Phone Trunk
- Genesys Cloud - CDM WebRTC Phone Trunk
- Indirect Tie TrunkBase for EdgeGroup
- PureCloud Voice - AWS
- Tie TrunkBase for EdgeGroup

### Important Notes

- These resources will be exported as a data resource
- Updates and modifications to this resource and its child dependencies will not be allowed through through this provider or the Genesys Cloud API
- This limitation is enforced by the Genesys Cloud API itself

This behavior ensures consistency with Genesys Cloud's management policies for managed telephony sites.
