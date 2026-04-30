package outbound_contact_list

import (
	"context"
	"encoding/csv"
	"encoding/json"
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
	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"
)

type outboundContactListRawResponse struct {
	PhoneColumns []outboundContactListRawPhoneColumn `json:"phoneColumns"`
	EmailColumns []outboundContactListRawEmailColumn `json:"emailColumns"`
}

type outboundContactListRawPhoneColumn struct {
	ColumnName             *string `json:"columnName"`
	VarType                *string `json:"type"`
	CallableTimeColumn     *string `json:"callableTimeColumn"`
	CallableTimeColumnName *string `json:"callableTimeColumnName"`
}

type outboundContactListRawEmailColumn struct {
	ColumnName                *string `json:"columnName"`
	VarType                   *string `json:"type"`
	ContactableTimeColumn     *string `json:"contactableTimeColumn"`
	ContactableTimeColumnName *string `json:"contactableTimeColumnName"`
}

func buildPhoneColumnTimeZoneNameIndex(raw []outboundContactListRawPhoneColumn) map[string]string {
	idx := make(map[string]string, len(raw))
	for _, c := range raw {
		if c.ColumnName == nil || c.VarType == nil {
			continue
		}
		if c.CallableTimeColumnName != nil && *c.CallableTimeColumnName != "" {
			idx[*c.ColumnName+"|"+*c.VarType] = *c.CallableTimeColumnName
			continue
		}
		// Fallback: some payloads may only include callableTimeColumn
		if c.CallableTimeColumn != nil && *c.CallableTimeColumn != "" {
			idx[*c.ColumnName+"|"+*c.VarType] = *c.CallableTimeColumn
		}
	}
	return idx
}

func buildEmailColumnTimeZoneNameIndex(raw []outboundContactListRawEmailColumn) map[string]string {
	idx := make(map[string]string, len(raw))
	for _, c := range raw {
		if c.ColumnName == nil || c.VarType == nil {
			continue
		}
		if c.ContactableTimeColumnName != nil && *c.ContactableTimeColumnName != "" {
			idx[*c.ColumnName+"|"+*c.VarType] = *c.ContactableTimeColumnName
			continue
		}
		if c.ContactableTimeColumn != nil && *c.ContactableTimeColumn != "" {
			idx[*c.ColumnName+"|"+*c.VarType] = *c.ContactableTimeColumn
		}
	}
	return idx
}

func parseOutboundContactListRaw(respBody []byte) (phoneIdx map[string]string, emailIdx map[string]string) {
	if len(respBody) == 0 {
		return nil, nil
	}
	var raw outboundContactListRawResponse
	if err := json.Unmarshal(respBody, &raw); err != nil {
		return nil, nil
	}
	return buildPhoneColumnTimeZoneNameIndex(raw.PhoneColumns), buildEmailColumnTimeZoneNameIndex(raw.EmailColumns)
}

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
		if callableTimeColumnName, ok := contactPhoneNumberColumnMap["callable_time_column_name"].(string); ok && callableTimeColumnName != "" {
			sdkContactPhoneNumberColumn.CallableTimeColumn = &callableTimeColumnName
		} else if callableTimeColumn := contactPhoneNumberColumnMap["callable_time_column"].(string); callableTimeColumn != "" {
			sdkContactPhoneNumberColumn.CallableTimeColumn = &callableTimeColumn
		}

		sdkContactPhoneNumberColumnSlice = append(sdkContactPhoneNumberColumnSlice, sdkContactPhoneNumberColumn)
	}
	return &sdkContactPhoneNumberColumnSlice
}

func flattenSdkOutboundContactListContactPhoneNumberColumnSlice(contactPhoneNumberColumns []platformclientv2.Contactphonenumbercolumn, callableTimeColumnNameIndex map[string]string) *schema.Set {
	if len(contactPhoneNumberColumns) == 0 {
		return nil
	}

	contactPhoneNumberColumnSet := schema.NewSet(hashOutboundContactListPhoneColumn, []interface{}{})
	for _, contactPhoneNumberColumn := range contactPhoneNumberColumns {
		contactPhoneNumberColumnMap := make(map[string]interface{})

		var key string
		if contactPhoneNumberColumn.ColumnName != nil {
			contactPhoneNumberColumnMap["column_name"] = *contactPhoneNumberColumn.ColumnName
			if contactPhoneNumberColumn.VarType != nil {
				key = *contactPhoneNumberColumn.ColumnName + "|" + *contactPhoneNumberColumn.VarType
			}
		}
		if contactPhoneNumberColumn.VarType != nil {
			contactPhoneNumberColumnMap["type"] = *contactPhoneNumberColumn.VarType
		}
		if contactPhoneNumberColumn.CallableTimeColumn != nil {
			// Keep legacy + new fields in sync while users migrate.
			contactPhoneNumberColumnMap["callable_time_column_name"] = *contactPhoneNumberColumn.CallableTimeColumn
			contactPhoneNumberColumnMap["callable_time_column"] = *contactPhoneNumberColumn.CallableTimeColumn
		} else if callableTimeColumnNameIndex != nil && key != "" {
			if v, ok := callableTimeColumnNameIndex[key]; ok && v != "" {
				contactPhoneNumberColumnMap["callable_time_column_name"] = v
				contactPhoneNumberColumnMap["callable_time_column"] = v
			}
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
		if contactableTimeColumnName, ok := contactEmailAddressColumnMap["contactable_time_column_name"].(string); ok && contactableTimeColumnName != "" {
			sdkContactEmailAddressColumn.ContactableTimeColumn = &contactableTimeColumnName
		} else if contactableTimeColumn, ok := contactEmailAddressColumnMap["contactable_time_column"].(string); ok && contactableTimeColumn != "" {
			sdkContactEmailAddressColumn.ContactableTimeColumn = &contactableTimeColumn
		}

		sdkContactEmailAddressColumnSlice = append(sdkContactEmailAddressColumnSlice, sdkContactEmailAddressColumn)
	}
	return &sdkContactEmailAddressColumnSlice
}

func flattenSdkOutboundContactListContactEmailAddressColumnSlice(contactEmailAddressColumns []platformclientv2.Emailcolumn, contactableTimeColumnNameIndex map[string]string) *schema.Set {
	if len(contactEmailAddressColumns) == 0 {
		return nil
	}

	contactEmailAddressColumnSet := schema.NewSet(hashOutboundContactListEmailColumn, []interface{}{})
	for _, contactEmailAddressColumn := range contactEmailAddressColumns {
		contactEmailAddressColumnMap := make(map[string]interface{})

		var key string
		if contactEmailAddressColumn.ColumnName != nil {
			contactEmailAddressColumnMap["column_name"] = *contactEmailAddressColumn.ColumnName
			if contactEmailAddressColumn.VarType != nil {
				key = *contactEmailAddressColumn.ColumnName + "|" + *contactEmailAddressColumn.VarType
			}
		}
		if contactEmailAddressColumn.VarType != nil {
			contactEmailAddressColumnMap["type"] = *contactEmailAddressColumn.VarType
		}
		if contactEmailAddressColumn.ContactableTimeColumn != nil {
			// Keep legacy + new fields in sync while users migrate.
			contactEmailAddressColumnMap["contactable_time_column_name"] = *contactEmailAddressColumn.ContactableTimeColumn
			contactEmailAddressColumnMap["contactable_time_column"] = *contactEmailAddressColumn.ContactableTimeColumn
		} else if contactableTimeColumnNameIndex != nil && key != "" {
			if v, ok := contactableTimeColumnNameIndex[key]; ok && v != "" {
				contactEmailAddressColumnMap["contactable_time_column_name"] = v
				contactEmailAddressColumnMap["contactable_time_column"] = v
			}
		}

		contactEmailAddressColumnSet.Add(contactEmailAddressColumnMap)
	}

	return contactEmailAddressColumnSet
}

func buildSdkOutboundContactListContactWhatsAppColumnSlice(contactWhatsAppColumn *schema.Set) *[]platformclientv2.Whatsappcolumn {
	if contactWhatsAppColumn == nil {
		return nil
	}
	sdkContactWhatsAppColumnSlice := make([]platformclientv2.Whatsappcolumn, 0)
	contactWhatsAppColumnList := contactWhatsAppColumn.List()
	for _, configWhatsAppColumn := range contactWhatsAppColumnList {
		var sdkContactWhatsAppColumn platformclientv2.Whatsappcolumn

		contactWhatsAppColumnMap, ok := configWhatsAppColumn.(map[string]interface{})
		if !ok {
			continue
		}

		// Safely handle column_name
		if columnName, ok := contactWhatsAppColumnMap["column_name"].(string); ok && columnName != "" {
			sdkContactWhatsAppColumn.ColumnName = &columnName
		}

		// Safely handle type
		if varType, ok := contactWhatsAppColumnMap["type"].(string); ok && varType != "" {
			sdkContactWhatsAppColumn.VarType = &varType
		}

		sdkContactWhatsAppColumnSlice = append(sdkContactWhatsAppColumnSlice, sdkContactWhatsAppColumn)
	}
	return &sdkContactWhatsAppColumnSlice
}

func flattenSdkOutboundContactListContactWhatsAppColumnSlice(contactWhatsAppColumns []platformclientv2.Whatsappcolumn) *schema.Set {
	if len(contactWhatsAppColumns) == 0 {
		return nil
	}

	contactWhatsAppColumnSet := schema.NewSet(schema.HashResource(outboundContactListWhatsAppColumnResource), []interface{}{})
	for _, contactWhatsAppColumn := range contactWhatsAppColumns {
		contactWhatsAppColumnMap := make(map[string]interface{})

		if contactWhatsAppColumn.ColumnName != nil {
			contactWhatsAppColumnMap["column_name"] = *contactWhatsAppColumn.ColumnName
		}
		if contactWhatsAppColumn.VarType != nil {
			contactWhatsAppColumnMap["type"] = *contactWhatsAppColumn.VarType
		}

		contactWhatsAppColumnSet.Add(contactWhatsAppColumnMap)
	}

	return contactWhatsAppColumnSet
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

	// Strip system-generated columns from the exported CSV to prevent CONTACT_COLUMNS_LIMIT_EXCEEDED
	// errors when importing into another org. The Genesys Cloud export API includes system metadata
	// columns (e.g. CallRecordLastAttempt, Callable, AutomaticTimeZone) that are not part of the
	// user-defined column_names and can cause the column limit to be exceeded.
	columnsToKeep := getListAttributeFromState(resource.State.Attributes, "column_names")
	if len(columnsToKeep) > 0 {
		columnsToKeep = append([]string{"inin-outbound-id"}, columnsToKeep...)
		if err := stripSystemColumnsFromCSV(fullCurrentPath, columnsToKeep); err != nil {
			return fmt.Errorf("failed to strip system columns from exported contact list CSV: %w", err)
		}
		log.Printf("Stripped system-generated columns from exported contact list %s CSV", contactListName)
	}

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
		callable_time_column_name = %s
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
		contactable_time_column_name = %s
	}
`, columnName, columnType, contactableTimeColumn)
}

// getListAttributeFromState reads a list attribute from Terraform's flat InstanceState format.
// In InstanceState, list attributes are stored as: attrName.# = "3", attrName.0 = "val1", etc.
func getListAttributeFromState(attrs map[string]string, attrName string) []string {
	countStr, ok := attrs[attrName+".#"]
	if !ok {
		return nil
	}
	count, err := strconv.Atoi(countStr)
	if err != nil || count == 0 {
		return nil
	}
	values := make([]string, 0, count)
	for i := 0; i < count; i++ {
		if val, ok := attrs[fmt.Sprintf("%s.%d", attrName, i)]; ok {
			values = append(values, val)
		}
	}
	return values
}

// stripSystemColumnsFromCSV reads a CSV file and rewrites it keeping only the columns specified in columnsToKeep.
func stripSystemColumnsFromCSV(filePath string, columnsToKeep []string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %w", err)
	}

	reader := csv.NewReader(f)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	allRecords, err := reader.ReadAll()
	f.Close()
	if err != nil {
		return fmt.Errorf("failed to read CSV file: %w", err)
	}

	if len(allRecords) == 0 {
		return nil
	}

	headers := allRecords[0]

	// Build a set of columns to keep for fast lookup
	keepSet := make(map[string]bool, len(columnsToKeep))
	for _, col := range columnsToKeep {
		keepSet[col] = true
	}

	// Find the indexes of columns to keep, preserving CSV column order
	var keepIndexes []int
	var rebuiltHeaders []string
	for i, header := range headers {
		// Some columns have extra whitespace or unicode that needs to be stripped
		// For example: "<feff>""inin-outbound-id"""
		header = util.StripInvisibleUnicodeFromString(header)

		// Some columns have extra quotes that need to be stripped (see above example).
		// The CSV writer will automagically put them back in and we have no control over
		// this behavior, but we don't want triple double-quote headers (i.e. """inin-outbound-id""")
		header = strings.Trim(header, "\"")

		if keepSet[header] {
			keepIndexes = append(keepIndexes, i)
			rebuiltHeaders = append(rebuiltHeaders, header)
		}
	}

	// If no columns were stripped, no need to rewrite
	if len(keepIndexes) == len(headers) {
		return nil
	}

	log.Printf("Stripping %d system-generated columns from CSV (keeping %d of %d)", len(headers)-len(keepIndexes), len(keepIndexes), len(headers))

	// Build stripped records
	strippedColumnsRecords := make([][]string, len(allRecords))

	// Set the header row with rebuilt headers
	strippedColumnsRecords[0] = rebuiltHeaders

	// Process data rows
	for i, record := range allRecords[1:len(allRecords)] {
		strippedRow := make([]string, len(keepIndexes))
		for j, idx := range keepIndexes {
			if idx < len(record) {
				strippedRow[j] = record[idx]
			}
		}
		strippedColumnsRecords[i+1] = strippedRow // Index is 1-indexed since we already wrote the header
	}

	// Write back to the same file
	out, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer out.Close()

	writer := csv.NewWriter(out)
	if err := writer.WriteAll(strippedColumnsRecords); err != nil {
		return fmt.Errorf("failed to write CSV file: %w", err)
	}

	return nil
}

func GenerateWhatsAppColumnsBlock(columnName, columnType string) string {
	return fmt.Sprintf(`
	whats_app_columns {
		column_name             = "%s"
		type                    = "%s"
	}
`, columnName, columnType)
}
