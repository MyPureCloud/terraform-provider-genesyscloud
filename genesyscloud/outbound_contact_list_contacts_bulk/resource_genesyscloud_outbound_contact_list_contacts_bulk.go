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

func createOutboundContactListBulkContacts(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	contactListId, contactsCount, diagErr := uploadOutboundContactListBulkContacts(ctx, d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Finished creating %d bulk contacts in contact list '%s'", contactsCount, contactListId)
	return readOutboundContactListBulkContacts(ctx, d, meta)
}

func readOutboundContactListBulkContacts(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	cp := getBulkContactsProxy(sdkConfig)
	contactListId := d.Get("contact_list_id").(string)
	if contactListId == "" {
		contactListId = getContactListIdFromResourceId(d.Id())
		d.Set("contact_list_id", contactListId)
	}
	contactList, contactListContactsCount, _, err := cp.readContactListAndRecordLengthById(ctx, contactListId)
	if err != nil {
		return diag.Errorf("Failed to read contact list and record length by ID: %v", err)
	}
	d.Set("record_count", contactListContactsCount)
	d.Set("contact_list_name", *contactList.Name)

	log.Printf("Read %d bulk contact records in contact list '%s'", contactListContactsCount, *contactList.Name)
	return nil
}

func deleteOutboundContactListBulkContacts(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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

func uploadOutboundContactListBulkContacts(ctx context.Context, d *schema.ResourceData, meta any) (string, int, diag.Diagnostics) {

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	cp := getBulkContactsProxy(sdkConfig)
	contactListContactsCount := 0

	filePath := d.Get("filepath").(string)
	filePathHash, err := getFileContentHash(filePath)
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

	if d.Get("file_content_hash") != filePathHash {
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
				return retry.RetryableError(fmt.Errorf("Number of records in the CSV file (%d) does not match the number of records in the contact list via the API (%d). Retrying.", csvRecordsCount, contactListContactsCount))
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
