package outbound_contact_list

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/files"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

func buildSdkOutboundContactListContactPhoneNumberColumnSlice(contactPhoneNumberColumn *schema.Set) *[]platformclientv2.Contactphonenumbercolumn {
	if contactPhoneNumberColumn == nil {
		return nil
	}
	sdkContactPhoneNumberColumnSlice := make([]platformclientv2.Contactphonenumbercolumn, 0)
	contactPhoneNumberColumnList := contactPhoneNumberColumn.List()
	for _, configPhoneColumn := range contactPhoneNumberColumnList {
		var sdkContactPhoneNumberColumn platformclientv2.Contactphonenumbercolumn
		contactPhoneNumberColumnMap := configPhoneColumn.(map[string]interface{})
		if columnName := contactPhoneNumberColumnMap["column_name"].(string); columnName != "" {
			sdkContactPhoneNumberColumn.ColumnName = &columnName
		}
		if varType := contactPhoneNumberColumnMap["type"].(string); varType != "" {
			sdkContactPhoneNumberColumn.VarType = &varType
		}
		if callableTimeColumn := contactPhoneNumberColumnMap["callable_time_column"].(string); callableTimeColumn != "" {
			sdkContactPhoneNumberColumn.CallableTimeColumn = &callableTimeColumn
		}

		sdkContactPhoneNumberColumnSlice = append(sdkContactPhoneNumberColumnSlice, sdkContactPhoneNumberColumn)
	}
	return &sdkContactPhoneNumberColumnSlice
}

func flattenSdkOutboundContactListContactPhoneNumberColumnSlice(contactPhoneNumberColumns []platformclientv2.Contactphonenumbercolumn) *schema.Set {
	if len(contactPhoneNumberColumns) == 0 {
		return nil
	}

	contactPhoneNumberColumnSet := schema.NewSet(schema.HashResource(outboundContactListContactPhoneNumberColumnResource), []interface{}{})
	for _, contactPhoneNumberColumn := range contactPhoneNumberColumns {
		contactPhoneNumberColumnMap := make(map[string]interface{})

		if contactPhoneNumberColumn.ColumnName != nil {
			contactPhoneNumberColumnMap["column_name"] = *contactPhoneNumberColumn.ColumnName
		}
		if contactPhoneNumberColumn.VarType != nil {
			contactPhoneNumberColumnMap["type"] = *contactPhoneNumberColumn.VarType
		}
		if contactPhoneNumberColumn.CallableTimeColumn != nil {
			contactPhoneNumberColumnMap["callable_time_column"] = *contactPhoneNumberColumn.CallableTimeColumn
		}

		contactPhoneNumberColumnSet.Add(contactPhoneNumberColumnMap)
	}

	return contactPhoneNumberColumnSet
}

func buildSdkOutboundContactListContactEmailAddressColumnSlice(contactEmailAddressColumn *schema.Set) *[]platformclientv2.Emailcolumn {
	if contactEmailAddressColumn == nil {
		return nil
	}
	sdkContactEmailAddressColumnSlice := make([]platformclientv2.Emailcolumn, 0)
	contactEmailAddressColumnList := contactEmailAddressColumn.List()
	for _, configEmailColumn := range contactEmailAddressColumnList {
		var sdkContactEmailAddressColumn platformclientv2.Emailcolumn

		contactEmailAddressColumnMap, ok := configEmailColumn.(map[string]interface{})
		if !ok {
			continue
		}

		// Safely handle column_name
		if columnName, ok := contactEmailAddressColumnMap["column_name"].(string); ok && columnName != "" {
			sdkContactEmailAddressColumn.ColumnName = &columnName
		}

		// Safely handle type
		if varType, ok := contactEmailAddressColumnMap["type"].(string); ok && varType != "" {
			sdkContactEmailAddressColumn.VarType = &varType
		}

		// Safely handle contactable_time_column
		if contactableTimeColumn, ok := contactEmailAddressColumnMap["contactable_time_column"].(string); ok && contactableTimeColumn != "" {
			sdkContactEmailAddressColumn.ContactableTimeColumn = &contactableTimeColumn
		}

		sdkContactEmailAddressColumnSlice = append(sdkContactEmailAddressColumnSlice, sdkContactEmailAddressColumn)
	}
	return &sdkContactEmailAddressColumnSlice
}

func flattenSdkOutboundContactListContactEmailAddressColumnSlice(contactEmailAddressColumns []platformclientv2.Emailcolumn) *schema.Set {
	if len(contactEmailAddressColumns) == 0 {
		return nil
	}

	contactEmailAddressColumnSet := schema.NewSet(schema.HashResource(outboundContactListEmailColumnResource), []interface{}{})
	for _, contactEmailAddressColumn := range contactEmailAddressColumns {
		contactEmailAddressColumnMap := make(map[string]interface{})

		if contactEmailAddressColumn.ColumnName != nil {
			contactEmailAddressColumnMap["column_name"] = *contactEmailAddressColumn.ColumnName
		}
		if contactEmailAddressColumn.VarType != nil {
			contactEmailAddressColumnMap["type"] = *contactEmailAddressColumn.VarType
		}
		if contactEmailAddressColumn.ContactableTimeColumn != nil {
			contactEmailAddressColumnMap["contactable_time_column"] = *contactEmailAddressColumn.ContactableTimeColumn
		}

		contactEmailAddressColumnSet.Add(contactEmailAddressColumnMap)
	}

	return contactEmailAddressColumnSet
}

func buildSdkOutboundContactListColumnDataTypeSpecifications(columnDataTypeSpecifications []interface{}) *[]platformclientv2.Columndatatypespecification {
	if len(columnDataTypeSpecifications) < 1 {
		return nil
	}

	sdkColumnDataTypeSpecificationsSlice := make([]platformclientv2.Columndatatypespecification, 0)

	for _, spec := range columnDataTypeSpecifications {
		if specMap, ok := spec.(map[string]interface{}); ok {
			var sdkColumnDataTypeSpecification platformclientv2.Columndatatypespecification
			if columnNameStr, ok := specMap["column_name"].(string); ok {
				sdkColumnDataTypeSpecification.ColumnName = &columnNameStr
			}
			if columnDataTypeStr, ok := specMap["column_data_type"].(string); ok && columnDataTypeStr != "" {
				sdkColumnDataTypeSpecification.ColumnDataType = &columnDataTypeStr
			}
			if minInt, ok := specMap["min"].(int); ok {
				sdkColumnDataTypeSpecification.Min = &minInt
			}
			if maxInt, ok := specMap["max"].(int); ok {
				sdkColumnDataTypeSpecification.Max = &maxInt
			}
			if maxLengthInt, ok := specMap["max_length"].(int); ok {
				sdkColumnDataTypeSpecification.MaxLength = &maxLengthInt
			}
			sdkColumnDataTypeSpecificationsSlice = append(sdkColumnDataTypeSpecificationsSlice, sdkColumnDataTypeSpecification)
		}
	}

	return &sdkColumnDataTypeSpecificationsSlice
}

func flattenSdkOutboundContactListColumnDataTypeSpecifications(columnDataTypeSpecifications []platformclientv2.Columndatatypespecification) []interface{} {
	if len(columnDataTypeSpecifications) == 0 {
		return nil
	}

	columnDataTypeSpecificationsSlice := make([]interface{}, 0)

	for _, s := range columnDataTypeSpecifications {
		columnDataTypeSpecification := make(map[string]interface{})
		columnDataTypeSpecification["column_name"] = *s.ColumnName

		if s.ColumnDataType != nil {
			columnDataTypeSpecification["column_data_type"] = *s.ColumnDataType
		}
		if s.Min != nil {
			columnDataTypeSpecification["min"] = *s.Min
		}
		if s.Max != nil {
			columnDataTypeSpecification["max"] = *s.Max
		}
		if s.MaxLength != nil {
			columnDataTypeSpecification["max_length"] = *s.MaxLength
		}

		columnDataTypeSpecificationsSlice = append(columnDataTypeSpecificationsSlice, columnDataTypeSpecification)
	}

	return columnDataTypeSpecificationsSlice
}

func ContactsExporterResolver(resourceId, exportDirectory, subDirectory string, configMap map[string]interface{}, meta interface{}, resource resourceExporter.ResourceInfo) error {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	cp := GetOutboundContactlistProxy(sdkConfig)

	contactListName := resource.BlockLabel
	contactListId := resource.State.Attributes["id"]
	exportFileName := fmt.Sprintf("%s.csv", contactListName)

	fullDirectoryPath := filepath.Join(exportDirectory, subDirectory)
	if err := os.MkdirAll(fullDirectoryPath, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", fullDirectoryPath, err)
	}

	ctx := context.Background()
	var exportUrl string
	diagErr := util.RetryWhen(util.IsStatus404, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Initiating contact list export for contact list %s", contactListName)
		resp, err := cp.initiateContactListContactsExport(ctx, contactListId)
		if err != nil {
			return resp, diag.FromErr(err)
		}
		return resp, nil
	}, 400)

	if diagErr != nil {
		return fmt.Errorf(`error initiating contact list export: %v`, diagErr)
	}

	retryAttempt := 1
	diagErr = util.RetryWhen(util.IsStatus404, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Waiting for signed export URL for contact list %s", contactListName)
		var err error
		var resp *platformclientv2.APIResponse
		exportUrl, resp, err = cp.getContactListContactsExportUrl(ctx, contactListId)
		if err == nil {
			return resp, nil
		}

		// not a retry error so don't sleep
		if !util.IsStatus404(resp) && !util.IsStatus400(resp) {
			return resp, diag.FromErr(err)
		}

		// Give the system time to generate the URL
		// Exponential backoff - 1st sleep: 2 seconds, 10th sleep: 1024 seconds (total: 2046 seconds/34 minutes)
		waitTime := time.Duration(2*math.Pow(2, float64(retryAttempt)-1)) * time.Second
		log.Printf("Sleeping for %f seconds before retrying", waitTime.Seconds())
		time.Sleep(waitTime)

		retryAttempt += 1
		return resp, diag.FromErr(err)
	}, http.StatusBadRequest)

	if diagErr != nil {
		return fmt.Errorf(`error retrieving signed export url for contact list: %v`, diagErr)
	}

	diagErr = util.RetryWhen(util.IsStatus404, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Downloading exported contacts for contact list %s", contactListName)
		resp, err := files.DownloadExportFileWithAccessToken(fullDirectoryPath, exportFileName, exportUrl, sdkConfig.AccessToken)
		if err != nil {
			return resp, diag.FromErr(err)
		}
		return resp, nil
	}, 400)
	if diagErr != nil {
		return fmt.Errorf(`error downloading exported contacts: %v`, diagErr)
	}

	fullCurrentPath := filepath.Join(fullDirectoryPath, exportFileName)
	fullRelativePath := filepath.Join(subDirectory, exportFileName)
	log.Printf("Saving exported contact list %s to file %s and updating state", contactListName, fullRelativePath)
	configMap["contacts_filepath"] = fullRelativePath
	configMap["contacts_id_name"] = "inin-outbound-id"

	// Remove read only attributes from the config file
	delete(configMap, "contacts_file_content_hash")
	delete(configMap, "contacts_record_count")
	hash, err := files.HashFileContent(ctx, fullCurrentPath, S3Enabled)
	if err != nil {
		log.Printf("Error calculating file content hash: %v", err)
		return err
	}
	resource.State.Attributes["contacts_file_content_hash"] = hash

	recordCount, err := files.GetCSVRecordCount(fullCurrentPath)
	if err != nil {
		log.Printf("Error getting CSV record count: %v", err)
		return err
	}
	resource.State.Attributes["contacts_record_count"] = strconv.Itoa(recordCount)

	resource.State.Attributes["contacts_filepath"] = fullRelativePath
	resource.State.Attributes["contacts_id_name"] = "inin-outbound-id"

	return nil
}

func GeneratePhoneColumnsBlock(columnName, columnType, callableTimeColumn string) string {
	return fmt.Sprintf(`
	phone_columns {
		column_name          = "%s"
		type                 = "%s"
		callable_time_column = %s
	}
`, columnName, columnType, callableTimeColumn)
}

func GenerateContactsFile(filepath, contactsIdName string) string {
	return fmt.Sprintf(`
	contacts_filepath = "%s"
	contacts_id_name = "%s"
	`, filepath, contactsIdName)
}

func GenerateOutboundContactList(
	resourceLabel string,
	name string,
	divisionId string,
	previewModeColumnName string,
	previewModeAcceptedValues []string,
	columnNames []string,
	automaticTimeZoneMapping string,
	zipCodeColumnName string,
	attemptLimitId string,
	nestedBlocks ...string) string {
	return fmt.Sprintf(`
resource "%s" "%s" {
	name                         = "%s"
	division_id                  = %s
	preview_mode_column_name     = %s
	preview_mode_accepted_values = [%s]
	column_names                 = [%s]
	automatic_time_zone_mapping  = %s
	zip_code_column_name         = %s
	attempt_limit_id             = %s
	%s
}
`, ResourceType, resourceLabel, name, divisionId, previewModeColumnName, strings.Join(previewModeAcceptedValues, ", "),
		strings.Join(columnNames, ", "), automaticTimeZoneMapping, zipCodeColumnName, attemptLimitId, strings.Join(nestedBlocks, "\n"))
}

func GeneratePhoneColumnsDataTypeSpecBlock(columnName, columnDataType, min, max, maxLength string) string {
	return fmt.Sprintf(`
	column_data_type_specifications {
		column_name      = %s
		column_data_type = %s
		min              = %s
		max              = %s
		max_length       = %s
	}
	`, columnName, columnDataType, min, max, maxLength)
}

func GenerateEmailColumnsBlock(columnName, columnType, contactableTimeColumn string) string {
	return fmt.Sprintf(`
	email_columns {
		column_name             = "%s"
		type                    = "%s"
		contactable_time_column = %s
	}
`, columnName, columnType, contactableTimeColumn)
}
