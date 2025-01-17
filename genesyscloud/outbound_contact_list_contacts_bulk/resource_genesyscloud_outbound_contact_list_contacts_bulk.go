package outbound_contact_list_contacts_bulk

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
)

func getAllContactLists(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	cp := getBulkContactsProxy(clientConfig)

	contactLists, resp, err := cp.getAllContactLists(ctx)
	if err != nil {
		msg := fmt.Sprintf("Failed to read all contact lists. Error: %v", err)
		if resp != nil {
			return nil, util.BuildAPIDiagnosticError(ResourceType, msg, resp)
		}
		return nil, util.BuildDiagnosticError(ResourceType, msg, err)
	}

	for _, contactList := range *contactLists {
		resources[buildBulkContactId(*contactList.Id)] = &resourceExporter.ResourceMeta{BlockLabel: *contactList.Name + "_contacts_bulk"}
	}

	return resources, nil
}

func createOutboundContactListContact(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	contactListId, contactsCount, diagErr := uploadOutboundContactListContacts(ctx, d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Finished creating %s bulk contacts in contact list '%s'", contactsCount, contactListId)
	return readOutboundContactListContact(ctx, d, meta)
}

func readOutboundContactListContact(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	cp := getBulkContactsProxy(sdkConfig)
	contactListId := d.Get("contact_list_id").(string)
	_, contactListContactsCount, _, err := cp.readContactListAndRecordLengthById(ctx, contactListId)
	if err != nil {
		return diag.Errorf("Failed to read contact list and record length by ID: %v", err)
	}
	d.Set("contact_list_contacts_count", contactListContactsCount)

	log.Printf("Read %s bulk contact records in contact list '%s'", contactListContactsCount, contactListId)
	return nil
}

func updateOutboundContactListContact(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {

	contactListId, contactsCount, diagErr := uploadOutboundContactListContacts(ctx, d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Finished updating %s bulk contacts in contact list '%s'", contactsCount, contactListId)
	return readOutboundContactListContact(ctx, d, meta)
}

func deleteOutboundContactListContact(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	cp := getBulkContactsProxy(sdkConfig)
	contactListId := d.Get("contact_list_id").(string)

	log.Printf("Clearing all bulk contacts from contact list '%s'", contactListId)

	cp.clearContactListBulkContacts(ctx, contactListId)

	return util.WithRetries(ctx, 60*time.Second, func() *retry.RetryError {
		log.Printf("Reading contacts for contact list '%s'", contactListId)
		_, contactListContactsCount, resp, err := cp.readContactListAndRecordLengthById(ctx, contactListId)
		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Contact list '%s' no longer exists", contactListId)
				return nil
			}
			msg := fmt.Sprintf("failed to read contact list '%s'. Error: %v", contactListId, err)
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, msg, resp))
		}
		if contactListContactsCount > 0 {
			msg := fmt.Sprintf("contact list '%s' still has %d contacts", contactListId, contactListContactsCount)
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, msg, resp))
		}
		return nil
	})
}

func uploadOutboundContactListContacts(ctx context.Context, d *schema.ResourceData, meta any) (string, int, diag.Diagnostics) {

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	cp := getBulkContactsProxy(sdkConfig)
	contactListContactsCount := 0

	filePath := d.Get("file_path").(string)
	filePathHash, err := fileContentHashReader(filePath)
	if err != nil {
		return "", 0, diag.Errorf("Failed to read file content hash: %v", err)
	}

	contactListId := d.Get("contact_list_id").(string)
	// contactListTemplateId := d.Get("contact_list_template_id").(string)
	contactIdName := d.Get("contact_id_name").(string)
	// listNamePrefix := d.Get("list_name_prefix").(string)
	// divisionIdForTargetContactLists := d.Get("division_id_for_target_contact_lists").(string)

	if filePath == "" {
		// Shouldn't happen because Terraform should detect this in the schema first
		return "", 0, diag.Errorf("File path is required")
	}

	if d.HasChange("file_content_hash") {
		csvRecordsCount, err := cp.getCSVRecordCount(filePath)

		if contactListId != "" {
			_, err := cp.uploadContactListBulkContacts(ctx, contactListId, filePath, contactIdName)
			if err != nil {
				return "", 0, diag.Errorf("Failed to upload contact list bulk contacts: %v", err)
			}
		}
		// if contactListTemplateId != "" {
		// 	_, err := cp.uploadContactListTemplateBulkContacts(ctx, contactListTemplateId, filePath, contactIdName, listNamePrefix, divisionIdForTargetContactLists)
		// 	if err != nil {
		// 		return diag.Errorf("Failed to upload contact list template bulk contacts: %v", err)
		// 	}
		// }
		// Validate number of records
		diagErr := util.WithRetries(ctx, 60*time.Second, func() *retry.RetryError {
			// Sleep for 5 seconds before (re)trying
			time.Sleep(5 * time.Second)

			_, contactListContactsCount, _, err = cp.readContactListAndRecordLengthById(ctx, contactListId)
			if err != nil {
				return retry.NonRetryableError(err)
			}
			if csvRecordsCount != contactListContactsCount {
				return retry.RetryableError(fmt.Errorf("Number of records in the CSV file (%s) does not match the number of records in the contact list via the API (%s). Retrying.", csvRecordsCount, contactListContactsCount))
			}
			return nil
		})
		if diagErr != nil {
			return contactListId, contactListContactsCount, diag.Errorf("Failed to validate number of records in the CSV file: %v", diagErr)
		}

		d.Set("file_content_hash", filePathHash)
		d.SetId(buildBulkContactId(contactListId))
	}
	return contactListId, contactListContactsCount, nil
}
