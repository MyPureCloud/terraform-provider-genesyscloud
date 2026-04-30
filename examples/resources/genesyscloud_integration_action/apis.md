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

### Static Data Actions Are Not Exported

When exporting integration actions via the `genesyscloud_tf_export` resource, **static (built-in) data actions are intentionally excluded from the exported configuration**. Static data actions are the pre-installed system actions that ship with each Genesys Cloud integration; their IDs are prefixed with `static` (for example, `static_e7b86b86-...`).

These actions are owned and managed by Genesys Cloud and cannot be created, updated, or deleted through the public Integration Actions API. Exporting them would therefore produce Terraform configuration that fails on apply, so the exporter skips them on purpose.

#### What this means for you

- Only custom integration actions that you (or your team) have created will appear in the exported `.tf`/`.tf.json` output.
- If a static data action is missing from your export, this is expected behavior — not a bug or a permissions issue.
- If you need to reference a static data action from another resource (for example, an Architect flow), reference it by its existing static ID directly rather than expecting it to be present in the exported configuration.