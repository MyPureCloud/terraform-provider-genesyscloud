* [GET /api/v2/integrations/actions](https://developer.genesys.cloud/api/rest/v2/integrations/#get-api-v2-integrations-actions)
* [POST /api/v2/integrations/actions](https://developer.genesys.cloud/api/rest/v2/integrations/#post-api-v2-integrations-actions)
* [GET /api/v2/integrations/actions/{actionId}](https://developer.genesys.cloud/api/rest/v2/integrations/#get-api-v2-integrations-actions--actionId-)
* [GET /api/v2/integrations/actions/{actionId}/templates/{fileName}](https://developer.genesys.cloud/api/rest/v2/integrations/#get-api-v2-integrations-actions--actionId--templates--fileName-)
* [PATCH /api/v2/integrations/actions/{actionId}](https://developer.genesys.cloud/api/rest/v2/integrations/#patch-api-v2-integrations-actions--actionId-)
* [DELETE /api/v2/integrations/actions/{actionId}](https://developer.genesys.cloud/api/rest/v2/integrations/#delete-api-v2-integrations-actions--actionId-)

## Function Configuration APIs

* [POST /api/v2/integrations/actions/{actionId}/draft/function/upload](https://developer.genesys.cloud/api/rest/v2/integrations/#post-api-v2-integrations-actions--actionId--draft-function-upload)
* [GET /api/v2/integrations/actions/{actionId}/draft/function](https://developer.genesys.cloud/api/rest/v2/integrations/#get-api-v2-integrations-actions--actionId--draft-function)
* [PUT /api/v2/integrations/actions/{actionId}/draft/function](https://developer.genesys.cloud/api/rest/v2/integrations/#put-api-v2-integrations-actions--actionId--draft-function)
* [POST /api/v2/integrations/actions/{actionId}/draft/publish](https://developer.genesys.cloud/api/rest/v2/integrations/#post-api-v2-integrations-actions--actionId--draft-publish)

## Export Behavior

### Static Data Actions Are Exported as Data Sources

When exporting integration actions via the `genesyscloud_tf_export` resource, **static (built-in) data actions are emitted as `data` blocks rather than `resource` blocks**. Static data actions are the pre-installed system actions that ship with each Genesys Cloud integration; their IDs are prefixed with `static` (for example, `static_e7b86b86-...`).

These actions are owned and managed by Genesys Cloud and cannot be created, updated, or deleted through the public Integration Actions API. Emitting them as managed resources would therefore produce Terraform configuration that fails on apply. Exporting them as data sources lets other resources (for example, Architect flows) reference them by name while leaving lifecycle management to Genesys Cloud.

#### What this means for you

- Custom integration actions that you (or your team) created continue to be exported as `resource "genesyscloud_integration_action"` blocks.
- Static (built-in) data actions are exported as `data "genesyscloud_integration_action"` blocks that look them up by `name` and `integration_id`.
- References to static data actions from other exported resources are automatically rewritten to use the generated data source (for example, `data.genesyscloud_integration_action.<label>.id`).
- The `integration_id` attribute on the data source is optional, but it is emitted during export to disambiguate static actions whose names may repeat across integration instances.