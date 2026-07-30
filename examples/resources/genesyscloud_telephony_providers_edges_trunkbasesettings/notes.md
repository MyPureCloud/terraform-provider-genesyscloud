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
