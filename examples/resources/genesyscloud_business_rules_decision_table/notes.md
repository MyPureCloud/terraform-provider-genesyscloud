## Migrating from nested `rows` to `rows_csv_filepath`

### Deprecation notice

The nested `rows` block is deprecated and will be removed in a later version. Prefer `rows_csv_filepath` (with computed `rows_csv_content_hash` and `rows_record_count`). Existing configs that still use `rows` continue to work until removal.

### Migration steps

1. Obtain a CSV of table rows (recommended: run `genesyscloud_tf_export` against the existing table — writes a Populated CSV under `rows/` with `rowId` stripped and sets `rows_csv_filepath` in the exported config), or author a CSV yourself.
2. CSV rules: headers must match the platform export shape (`schema_property_key::Comparator` for inputs, bare `schema_property_key` for outputs); do **not** include a `rowId` column in the on-disk file — the import API (Replace mode) requires a `rowId` header with empty cell values, and the provider reinjects that column on upload (and strips it on export); `stringList` cells use `||` as the item delimiter; queue and other platform object cells use friendly names.
3. In the decision table resource, remove the nested `rows` blocks and set:

   ```hcl
   rows_csv_filepath = "${path.module}/rows/example.csv"
   ```

4. Run `terraform plan` / `terraform apply`. The resource ID is unchanged. Both the first apply after switching and later CSV edits (detected via content hash) run a Replace CSV import, then publish a new version.
