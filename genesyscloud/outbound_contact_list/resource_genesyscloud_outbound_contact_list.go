package outbound_contact_list

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/files"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

func getAllOutboundContactLists(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := GetOutboundContactlistProxy(clientConfig)

	contactLists, resp, getErr := proxy.GetAllOutboundContactlist(ctx)
	if getErr != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get contact lists error: %s", getErr), resp)
	}

	for _, contactList := range *contactLists {
		resources[*contactList.Id] = &resourceExporter.ResourceMeta{BlockLabel: *contactList.Name}
	}

	return resources, nil
}

func createOutboundContactList(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	columnNames := lists.InterfaceListToStrings(d.Get("column_names").([]interface{}))
	previewModeColumnName := d.Get("preview_mode_column_name").(string)
	previewModeAcceptedValues := lists.InterfaceListToStrings(d.Get("preview_mode_accepted_values").([]interface{}))
	automaticTimeZoneMapping := d.Get("automatic_time_zone_mapping").(bool)
	zipCodeColumnName := d.Get("zip_code_column_name").(string)
	trimWhitespace := d.Get("trim_whitespace").(bool)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := GetOutboundContactlistProxy(sdkConfig)

	sdkContactList := platformclientv2.Contactlist{
		Division:                     util.BuildSdkDomainEntityRef(d, "division_id"),
		ColumnNames:                  &columnNames,
		PhoneColumns:                 buildSdkOutboundContactListContactPhoneNumberColumnSlice(d.Get("phone_columns").(*schema.Set)),
		EmailColumns:                 buildSdkOutboundContactListContactEmailAddressColumnSlice(d.Get("email_columns").(*schema.Set)),
		PreviewModeAcceptedValues:    &previewModeAcceptedValues,
		AttemptLimits:                util.BuildSdkDomainEntityRef(d, "attempt_limit_id"),
		AutomaticTimeZoneMapping:     &automaticTimeZoneMapping,
		ColumnDataTypeSpecifications: buildSdkOutboundContactListColumnDataTypeSpecifications(d.Get("column_data_type_specifications").([]interface{})),
		TrimWhitespace:               &trimWhitespace,
	}

	if name != "" {
		sdkContactList.Name = &name
	}
	if previewModeColumnName != "" {
		sdkContactList.PreviewModeColumnName = &previewModeColumnName
	}
	if zipCodeColumnName != "" {
		sdkContactList.ZipCodeColumnName = &zipCodeColumnName
	}

	log.Printf("Creating Outbound Contact List %s", name)
	outboundContactList, resp, err := proxy.createOutboundContactlist(ctx, &sdkContactList)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create Outbound Contact List %s error: %s", name, err), resp)
	}

	d.SetId(*outboundContactList.Id)

	log.Printf("Created Outbound Contact List %s %s", name, *outboundContactList.Id)

	diagErr := uploadOutboundContactListBulkContacts(ctx, d, meta)
	if diagErr != nil {
		return diagErr
	}

	return readOutboundContactList(ctx, d, meta)
}

func updateOutboundContactList(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	columnNames := lists.InterfaceListToStrings(d.Get("column_names").([]interface{}))
	previewModeColumnName := d.Get("preview_mode_column_name").(string)
	previewModeAcceptedValues := lists.InterfaceListToStrings(d.Get("preview_mode_accepted_values").([]interface{}))
	automaticTimeZoneMapping := d.Get("automatic_time_zone_mapping").(bool)
	zipCodeColumnName := d.Get("zip_code_column_name").(string)
	trimWhitespace := d.Get("trim_whitespace").(bool)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := GetOutboundContactlistProxy(sdkConfig)

	sdkContactList := platformclientv2.Contactlist{
		Division:                     util.BuildSdkDomainEntityRef(d, "division_id"),
		ColumnNames:                  &columnNames,
		PhoneColumns:                 buildSdkOutboundContactListContactPhoneNumberColumnSlice(d.Get("phone_columns").(*schema.Set)),
		EmailColumns:                 buildSdkOutboundContactListContactEmailAddressColumnSlice(d.Get("email_columns").(*schema.Set)),
		PreviewModeAcceptedValues:    &previewModeAcceptedValues,
		AttemptLimits:                util.BuildSdkDomainEntityRef(d, "attempt_limit_id"),
		AutomaticTimeZoneMapping:     &automaticTimeZoneMapping,
		ColumnDataTypeSpecifications: buildSdkOutboundContactListColumnDataTypeSpecifications(d.Get("column_data_type_specifications").([]interface{})),
		TrimWhitespace:               &trimWhitespace,
	}

	if name != "" {
		sdkContactList.Name = &name
	}
	if previewModeColumnName != "" {
		sdkContactList.PreviewModeColumnName = &previewModeColumnName
	}
	if zipCodeColumnName != "" {
		sdkContactList.ZipCodeColumnName = &zipCodeColumnName
	}

	log.Printf("Updating Outbound Contact List %s", name)
	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {

		_, resp, updateErr := proxy.updateOutboundContactlist(ctx, d.Id(), &sdkContactList)
		if updateErr != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update Outbound contact list %s error: %s", name, updateErr), resp)
		}
		return nil, nil
	})
	if diagErr != nil {
		return diagErr
	}

	diagErr = uploadOutboundContactListBulkContacts(ctx, d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated Outbound Contact List %s", name)
	return readOutboundContactList(ctx, d, meta)
}

func readOutboundContactList(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := GetOutboundContactlistProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceOutboundContactList(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading Outbound Contact List %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		sdkContactList, resp, getErr := proxy.GetOutboundContactlistById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read Outbound Contact List %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read Outbound Contact List %s | error: %s", d.Id(), getErr), resp))
		}

		if sdkContactList.Name != nil {
			_ = d.Set("name", *sdkContactList.Name)
		}
		if sdkContactList.Division != nil && sdkContactList.Division.Id != nil {
			_ = d.Set("division_id", *sdkContactList.Division.Id)
		}
		if sdkContactList.ColumnNames != nil {
			_ = d.Set("column_names", *sdkContactList.ColumnNames)
		}
		if sdkContactList.PhoneColumns != nil {
			_ = d.Set("phone_columns", flattenSdkOutboundContactListContactPhoneNumberColumnSlice(*sdkContactList.PhoneColumns))
		}
		if sdkContactList.EmailColumns != nil {
			_ = d.Set("email_columns", flattenSdkOutboundContactListContactEmailAddressColumnSlice(*sdkContactList.EmailColumns))
		}
		if sdkContactList.PreviewModeColumnName != nil {
			_ = d.Set("preview_mode_column_name", *sdkContactList.PreviewModeColumnName)
		}
		if sdkContactList.PreviewModeAcceptedValues != nil {
			_ = d.Set("preview_mode_accepted_values", *sdkContactList.PreviewModeAcceptedValues)
		}
		if sdkContactList.AttemptLimits != nil && sdkContactList.AttemptLimits.Id != nil {
			_ = d.Set("attempt_limit_id", *sdkContactList.AttemptLimits.Id)
		}
		if sdkContactList.AutomaticTimeZoneMapping != nil {
			_ = d.Set("automatic_time_zone_mapping", *sdkContactList.AutomaticTimeZoneMapping)
		}
		if sdkContactList.ZipCodeColumnName != nil {
			_ = d.Set("zip_code_column_name", *sdkContactList.ZipCodeColumnName)
		}
		if sdkContactList.ColumnDataTypeSpecifications != nil {
			_ = d.Set("column_data_type_specifications", flattenSdkOutboundContactListColumnDataTypeSpecifications(*sdkContactList.ColumnDataTypeSpecifications))
		}
		if sdkContactList.TrimWhitespace != nil {
			_ = d.Set("trim_whitespace", *sdkContactList.TrimWhitespace)
		}

		if sdkContactList.Id != nil {
			contactListRecordsCount, _, err := proxy.getOutboundContactlistContactRecordLength(ctx, *sdkContactList.Id)
			if err != nil {
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read Outbound Contact List's records %s | error: %s", d.Id(), err), resp))
			}
			d.Set("contacts_record_count", contactListRecordsCount)
		}

		log.Printf("Read Outbound Contact List %s %s", d.Id(), *sdkContactList.Name)
		return cc.CheckState(d)
	})
}

func deleteOutboundContactList(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := GetOutboundContactlistProxy(sdkConfig)

	diagErr := util.RetryWhen(util.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting Outbound Contact List")
		resp, err := proxy.deleteOutboundContactlist(ctx, d.Id())
		if err != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete Outbound Contact List %s error: %s", d.Id(), err), resp)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.GetOutboundContactlistById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				// Outbound Contact List deleted
				log.Printf("Deleted Outbound Contact List %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error deleting Outbound Contact List %s | error: %s", d.Id(), err), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Outbound Contact List %s still exists", d.Id()), resp))
	})
}

func uploadOutboundContactListBulkContacts(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	filePath := d.Get("contacts_filepath").(string)
	if filePath == "" {
		return nil
	}

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	cp := GetOutboundContactlistProxy(sdkConfig)

	filePathHash, err := files.HashFileContent(ctx, filePath, S3Enabled)
	if err != nil {
		return diag.Errorf("Failed to read file content hash: %v", err)
	}

	contactListId := d.Id()
	contactListName := d.Get("name").(string)
	contactsIdName := d.Get("contacts_id_name").(string)

	if d.Get("contacts_file_content_hash") == filePathHash {
		return nil
	}

	csvRecordsCount, err := files.GetCSVRecordCount(filePath)
	if err != nil {
		return diag.Errorf("Failed to get CSV record count: %v", err)
	}

	log.Printf("Clearing existing contacts on contact list %s in preparation for updating the latest contacts", contactListName)
	resp, err := cp.clearContactListContacts(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to clear contacts on contact list %s error: %s", contactListName, err), resp)
	}

	_, diagErr := validateContactsRecordCount(ctx, cp, d.Id(), 0)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Uploading %d contact records to %s contact list", csvRecordsCount, contactListName)

	if contactListId != "" {
		_, err := cp.uploadContactListBulkContacts(ctx, contactListId, filePath, contactsIdName)
		if err != nil {
			return diag.Errorf("Failed to upload contact list bulk contacts: %v", err)
		}
	}

	contactCount, diagErr := validateContactsRecordCount(ctx, cp, contactListId, csvRecordsCount)
	if diagErr.HasError() {
		return diagErr
	}
	d.Set("contacts_file_content_hash", filePathHash)
	d.Set("contacts_record_count", contactCount)

	return nil
}

// Validate number of contact records in a contact list
func validateContactsRecordCount(ctx context.Context, cp *OutboundContactlistProxy, contactListId string, expectedRecordCount int) (recordCount int, err diag.Diagnostics) {
	contactListContactsCount := 0
	diagErr := util.WithRetries(ctx, 60*time.Second, func() *retry.RetryError {

		// Sleep for 5 seconds before (re)trying as per documentation
		// https://developer.genesys.cloud/routing/outbound/contactmanagement#manipulate-contact-list
		time.Sleep(5 * time.Second)

		contactListContactsCount, _, err := cp.getOutboundContactlistContactRecordLength(ctx, contactListId)
		if err != nil {
			return retry.NonRetryableError(err)
		}
		if expectedRecordCount != contactListContactsCount {
			return retry.RetryableError(fmt.Errorf("Number of records in the CSV file (%d) does not match the number of records in the contact list via the API (%d). Retrying.", expectedRecordCount, contactListContactsCount))
		}
		return nil
	})
	if diagErr != nil {
		return contactListContactsCount, diag.Errorf("Failed to validate number of records in the CSV file: %v", diagErr)
	}
	return contactListContactsCount, nil
}
