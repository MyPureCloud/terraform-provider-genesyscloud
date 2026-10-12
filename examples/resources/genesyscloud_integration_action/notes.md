## Export Behavior

### Function Data Action Zip Files Cannot Be Exported

Genesys Cloud does not allow downloading uploaded function zip files (API or UI). See [Limitations of the Genesys Cloud Function data actions integration](https://help.genesys.cloud/articles/limitations-of-the-genesys-cloud-function-data-actions-integration/).

When you export `genesyscloud_integration_action` resources that use Function Data Actions, `function_config.file_path` is emitted as a Terraform variable (the same pattern the legacy Architect flow exporter uses for flow YAML that cannot be retrieved). A `variable` block and `terraform.tfvars` placeholder are generated. Set the variable to a local or S3 zip path before plan or apply. The zip binary itself is never written to the export directory.

### Static Data Actions Are Exported as Data Sources

When exporting integration actions via the `genesyscloud_tf_export` resource, **static (built-in) data actions are emitted as `data` blocks rather than `resource` blocks**. Static data actions are the pre-installed system actions that ship with each Genesys Cloud integration; their IDs are prefixed with `static` (for example, `static_e7b86b86-...`).

These actions are owned and managed by Genesys Cloud and cannot be created, updated, or deleted through the public Integration Actions API. Emitting them as managed resources would therefore produce Terraform configuration that fails on apply. Exporting them as data sources lets other resources (for example, Architect flows) reference them by name while leaving lifecycle management to Genesys Cloud.

#### What this means for you

- Custom integration actions that you (or your team) created continue to be exported as `resource "genesyscloud_integration_action"` blocks.
- Static (built-in) data actions are exported as `data "genesyscloud_integration_action"` blocks that look them up by `name` and `integration_id`.
- References to static data actions from other exported resources are automatically rewritten to use the generated data source (for example, `data.genesyscloud_integration_action.<label>.id`).
- The `integration_id` attribute on the data source is optional, but it is emitted during export to disambiguate static actions whose names may repeat across integration instances.
