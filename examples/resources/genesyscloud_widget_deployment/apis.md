* [GET /api/v2/widgets/deployments](https://developer.genesys.cloud/api/rest/v2/widgets/#get-api-v2-widgets-deployments)
* [GET /api/v2/widgets/deployments/{deploymentId}](https://developer.genesys.cloud/api/rest/v2/widgets/#get-api-v2-widgets-deployments--deploymentId-)
* [POST /api/v2/widgets/deployments](https://developer.genesys.cloud/api/rest/v2/widgets/#post-api-v2-widgets-deployments)
* [PUT /api/v2/widgets/deployments/{deploymentId}](https://developer.genesys.cloud/api/rest/v2/widgets/#put-api-v2-widgets-deployments--deploymentId-)
* [DELETE /api/v2/widgets/deployments/{deploymentId}](https://developer.genesys.cloud/api/rest/v2/widgets/#delete-api-v2-widgets-deployments--deploymentId-)

## Migrating from genesyscloud_widget_deployment

### Deprecation Notice

The `genesyscloud_widget_deployment` resource is deprecated and will be removed in a future version due to this functionality being sunset in Genesys Cloud API.

### Migration Steps

1. Remove any `genesyscloud_widget_deployment` resources from your Terraform configuration.
