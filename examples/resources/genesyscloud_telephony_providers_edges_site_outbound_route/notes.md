#### Compatibility Note

In versions 1.39.0 to 1.48.0 of the provider, this resource was constructed with a different structure. The current version introduces structural changes that are not backwards compatible with those earlier versions.

If you're upgrading from an earlier version, please be aware of these structural changes and consult these examples on how to migrate your configuration.

## Export Behavior

When exporting this resource, please be aware of the following behavior:

If the associated Genesys Cloud Telephony Site is configured as a `managed` resource:

- This resource will be exported as a data resource
- Updates and modifications to this resource and its child dependencies will not be allowed through this provider or the Genesys Cloud API
- This limitation is enforced by the Genesys Cloud API itself

This behavior ensures consistency with Genesys Cloud's management policies for managed telephony sites.
