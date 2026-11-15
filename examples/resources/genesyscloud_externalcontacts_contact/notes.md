## Exported block labels

By default, exported external contact block labels include the external organization name, followed by the contact's last and first names. This preserves existing block labels but requires an additional organization API request for each contact associated with an external organization.

Set the `GENESYSCLOUD_EXTERNAL_CONTACTS_SIMPLE_BLOCK_LABEL` environment variable to omit the organization name and avoid those additional API requests. When enabled, block labels use the contact's last and first names. The variable's value is ignored; its presence enables the simpler labels.