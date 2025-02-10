- [POST /api/v2/outbound/contactlists/{contactListId}/contacts](https://developer.genesys.cloud/devapps/api-explorer#post-api-v2-outbound-contactlists--contactListId--contacts)
- [GET /api/v2/outbound/contactlists/{contactListId}/contacts/{contactId}](https://developer.genesys.cloud/devapps/api-explorer#get-api-v2-outbound-contactlists--contactListId--contacts--contactId-)
- [PUT /api/v2/outbound/contactlists/{contactListId}/contacts/{contactId}](https://developer.genesys.cloud/devapps/api-explorer#put-api-v2-outbound-contactlists--contactListId--contacts--contactId-)
- [DELETE /api/v2/outbound/contactlists/{contactListId}/contacts/{contactId}](https://developer.genesys.cloud/devapps/api-explorer#delete-api-v2-outbound-contactlists--contactListId--contacts--contactId-)

## Migrating from genesyscloud_outbound_contact_list_contact

### Deprecation Notice

The `genesyscloud_outbound_contact_list_contact` resource is deprecated and will be removed in a future version. Instead, use the `contacts_filepath` and `contacts_id_name` attributes in the `genesyscloud_outbound_contact_list` resource.

### Note About Exporter

The exporter functionality has been removed from this resource in favor of the `genesyscloud_outbound_contact_list` resource's built-in bulk handling of contacts via CSV exports. Contacts will now be exported within the CSV file output from the `genesyscloud_outbound_contact_list` and not be exported via this resource due to performance and scalability limitations with this resource.

### Migration Steps

1. Remove any `genesyscloud_outbound_contact_list_contact` resources from your Terraform configuration and add them to a `contacts.csv` file
2. Update your `genesyscloud_outbound_contact_list` resource to include:

   ```hcl
   resource "genesyscloud_outbound_contact_list" "example" {
     name = "Example Contact List"
     # ... other existing configuration ...

     contacts_filepath = "path/to/your/contacts.csv"
     contacts_id_name = "contact_id_column"
   }
   ```
3. Ensure your CSV file contains all required columns defined in `column_names`
4. Run `terraform plan` to verify the changes
