## Note on Resource Type Name

This resource was previously available as `genesyscloud_intent_category` in earlier versions of the provider. It has been renamed in this version to `genesyscloud_intents_categories` to align with the Genesys Cloud API naming conventions. If you are upgrading from a prior version, update your configuration files and run `terraform state mv` to migrate existing state entries.
