package outbound_contact_list

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

var (
	outboundContactListContactPhoneNumberColumnResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`column_name`: {
				Description: `The name of the phone column.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`type`: {
				Description: `Indicates the type of the phone column. For example, 'cell' or 'home'.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`callable_time_column`: {
				Description: `A column that indicates the timezone to use for a given contact when checking callable times. Not allowed if 'automaticTimeZoneMapping' is set to true.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}

	outboundContactListEmailColumnResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`column_name`: {
				Description: `The name of the email column.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`type`: {
				Description: `Indicates the type of the email column. For example, 'work' or 'personal'.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`contactable_time_column`: {
				Description: `A column that indicates the timezone to use for a given contact when checking contactable times.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}

	outboundContactListColumnDataTypeSpecification = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`column_name`: {
				Description: `The column name of a column selected for dynamic queueing.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`column_data_type`: {
				Description:  `The data type of the column selected for dynamic queueing (TEXT, NUMERIC or TIMESTAMP)`,
				Optional:     true,
				Computed:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"TEXT", "NUMERIC", "TIMESTAMP"}, false),
			},
			`min`: {
				Description: `The minimum length of the numeric column selected for dynamic queueing.`,
				Optional:    true,
				Type:        schema.TypeInt,
			},
			`max`: {
				Description: `The maximum length of the numeric column selected for dynamic queueing.`,
				Optional:    true,
				Type:        schema.TypeInt,
			},
			`max_length`: {
				Description: `The maximum length of the text column selected for dynamic queueing.`,
				Required:    true,
				Type:        schema.TypeInt,
			},
		},
	}
)

func getAllOutboundContactLists(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	outboundAPI := platformclientv2.NewOutboundApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		contactListConfigs, _, getErr := outboundAPI.GetOutboundContactlists(false, false, pageSize, pageNum, true, "", "", []string{}, []string{}, "", "")
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of contact list configs: %v", getErr)
		}

		if contactListConfigs.Entities == nil || len(*contactListConfigs.Entities) == 0 {
			break
		}

		for _, contactListConfig := range *contactListConfigs.Entities {
			resources[*contactListConfig.Id] = &resourceExporter.ResourceMeta{Name: *contactListConfig.Name}
		}
	}

	return resources, nil
}

func OutboundContactListExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: gcloud.GetAllWithPooledClient(getAllOutboundContactLists),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"attempt_limit_id": {RefType: "genesyscloud_outbound_attempt_limit"},
			"division_id":      {RefType: "genesyscloud_auth_division"},
		},
	}
}

func ResourceOutboundContactList() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud Outbound Contact List`,

		CreateContext: gcloud.CreateWithPooledClient(createOutboundContactList),
		ReadContext:   gcloud.ReadWithPooledClient(readOutboundContactList),
		UpdateContext: gcloud.UpdateWithPooledClient(updateOutboundContactList),
		DeleteContext: gcloud.DeleteWithPooledClient(deleteOutboundContactList),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The name for the contact list.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`division_id`: {
				Description: `The division this entity belongs to.`,
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeString,
			},
			`column_names`: {
				Description: `The names of the contact data columns. Changing the column_names attribute will cause the outboundcontact_list object to be dropped and recreated with a new ID`,
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			`phone_columns`: {
				Description: `Indicates which columns are phone numbers. Changing the phone_columns attribute will cause the outboundcontact_list object to be dropped and recreated with a new ID. Required if email_columns is empty`,
				Optional:    true,
				ForceNew:    true,
				Type:        schema.TypeSet,
				Elem:        outboundContactListContactPhoneNumberColumnResource,
			},
			`email_columns`: {
				Description: `Indicates which columns are email addresses. Changing the email_columns attribute will cause the outboundcontact_list object to be dropped and recreated with a new ID. Required if phone_columns is empty`,
				Optional:    true,
				ForceNew:    true,
				Type:        schema.TypeSet,
				Elem:        outboundContactListEmailColumnResource,
			},
			`preview_mode_column_name`: {
				Description: `A column to check if a contact should always be dialed in preview mode.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`preview_mode_accepted_values`: {
				Description: `The values in the previewModeColumnName column that indicate a contact should always be dialed in preview mode.`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			`attempt_limit_id`: {
				Description: `Attempt Limit for this ContactList.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`automatic_time_zone_mapping`: {
				Description: `Indicates if automatic time zone mapping is to be used for this ContactList. Changing the automatic_time_zone_mappings attribute will cause the outboundcontact_list object to be dropped and recreated with a new ID`,
				Optional:    true,
				ForceNew:    true,
				Type:        schema.TypeBool,
			},
			`zip_code_column_name`: {
				Description: `The name of contact list column containing the zip code for use with automatic time zone mapping. Only allowed if 'automaticTimeZoneMapping' is set to true. Changing the zip_code_column_name attribute will cause the outboundcontact_list object to be dropped and recreated with a new ID`,
				Optional:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
			},
			`column_data_type_specifications`: {
				Description: `The settings of the columns selected for dynamic queueing. If updated, the contact list is dropped and recreated with a new ID`,
				Optional:    true,
				ForceNew:    true,
				Type:        schema.TypeList,
				Elem:        outboundContactListColumnDataTypeSpecification,
			},
		},
	}
}

func createOutboundContactList(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	columnNames := lists.InterfaceListToStrings(d.Get("column_names").([]interface{}))
	previewModeColumnName := d.Get("preview_mode_column_name").(string)
	previewModeAcceptedValues := lists.InterfaceListToStrings(d.Get("preview_mode_accepted_values").([]interface{}))
	automaticTimeZoneMapping := d.Get("automatic_time_zone_mapping").(bool)
	zipCodeColumnName := d.Get("zip_code_column_name").(string)

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	sdkContactList := platformclientv2.Contactlist{
		Division:                     gcloud.BuildSdkDomainEntityRef(d, "division_id"),
		ColumnNames:                  &columnNames,
		PhoneColumns:                 buildSdkOutboundContactListContactPhoneNumberColumnSlice(d.Get("phone_columns").(*schema.Set)),
		EmailColumns:                 buildSdkOutboundContactListContactEmailAddressColumnSlice(d.Get("email_columns").(*schema.Set)),
		PreviewModeAcceptedValues:    &previewModeAcceptedValues,
		AttemptLimits:                gcloud.BuildSdkDomainEntityRef(d, "attempt_limit_id"),
		AutomaticTimeZoneMapping:     &automaticTimeZoneMapping,
		ColumnDataTypeSpecifications: buildSdkOutboundContactListColumnDataTypeSpecifications(d.Get("column_data_type_specifications").([]interface{})),
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
	outboundContactList, _, err := outboundApi.PostOutboundContactlists(sdkContactList)
	if err != nil {
		return diag.Errorf("Failed to create Outbound Contact List %s: %s", name, err)
	}

	d.SetId(*outboundContactList.Id)

	log.Printf("Created Outbound Contact List %s %s", name, *outboundContactList.Id)
	return readOutboundContactList(ctx, d, meta)
}

func updateOutboundContactList(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	columnNames := lists.InterfaceListToStrings(d.Get("column_names").([]interface{}))
	previewModeColumnName := d.Get("preview_mode_column_name").(string)
	previewModeAcceptedValues := lists.InterfaceListToStrings(d.Get("preview_mode_accepted_values").([]interface{}))
	automaticTimeZoneMapping := d.Get("automatic_time_zone_mapping").(bool)
	zipCodeColumnName := d.Get("zip_code_column_name").(string)

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	sdkContactList := platformclientv2.Contactlist{
		Division:                     gcloud.BuildSdkDomainEntityRef(d, "division_id"),
		ColumnNames:                  &columnNames,
		PhoneColumns:                 buildSdkOutboundContactListContactPhoneNumberColumnSlice(d.Get("phone_columns").(*schema.Set)),
		EmailColumns:                 buildSdkOutboundContactListContactEmailAddressColumnSlice(d.Get("email_columns").(*schema.Set)),
		PreviewModeAcceptedValues:    &previewModeAcceptedValues,
		AttemptLimits:                gcloud.BuildSdkDomainEntityRef(d, "attempt_limit_id"),
		AutomaticTimeZoneMapping:     &automaticTimeZoneMapping,
		ColumnDataTypeSpecifications: buildSdkOutboundContactListColumnDataTypeSpecifications(d.Get("column_data_type_specifications").([]interface{})),
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
	diagErr := gcloud.RetryWhen(gcloud.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current Outbound Contact list version
		outboundContactList, resp, getErr := outboundApi.GetOutboundContactlist(d.Id(), false, false)
		if getErr != nil {
			return resp, diag.Errorf("Failed to read Outbound Contact List %s: %s", d.Id(), getErr)
		}
		sdkContactList.Version = outboundContactList.Version
		outboundContactList, _, updateErr := outboundApi.PutOutboundContactlist(d.Id(), sdkContactList)
		if updateErr != nil {
			return resp, diag.Errorf("Failed to update Outbound Contact List %s: %s", name, updateErr)
		}
		return nil, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated Outbound Contact List %s", name)
	return readOutboundContactList(ctx, d, meta)
}

func readOutboundContactList(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	log.Printf("Reading Outbound Contact List %s", d.Id())

	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		sdkContactList, resp, getErr := outboundApi.GetOutboundContactlist(d.Id(), false, false)
		if getErr != nil {
			if gcloud.IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("failed to read Outbound Contact List %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("failed to read Outbound Contact List %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceOutboundContactList())

		if sdkContactList.Name != nil {
			_ = d.Set("name", *sdkContactList.Name)
		}
		if sdkContactList.Division != nil && sdkContactList.Division.Id != nil {
			_ = d.Set("division_id", *sdkContactList.Division.Id)
		}
		if sdkContactList.ColumnNames != nil {
			var columnNames []string
			for _, name := range *sdkContactList.ColumnNames {
				columnNames = append(columnNames, name)
			}
			_ = d.Set("column_names", columnNames)
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
			var acceptedValues []string
			for _, val := range *sdkContactList.PreviewModeAcceptedValues {
				acceptedValues = append(acceptedValues, val)
			}
			_ = d.Set("preview_mode_accepted_values", acceptedValues)
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

		log.Printf("Read Outbound Contact List %s %s", d.Id(), *sdkContactList.Name)
		return cc.CheckState()
	})
}

func deleteOutboundContactList(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	diagErr := gcloud.RetryWhen(gcloud.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting Outbound Contact List")
		resp, err := outboundApi.DeleteOutboundContactlist(d.Id())
		if err != nil {
			return resp, diag.Errorf("Failed to delete Outbound Contact List: %s", err)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return gcloud.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := outboundApi.GetOutboundContactlist(d.Id(), false, false)
		if err != nil {
			if gcloud.IsStatus404(resp) {
				// Outbound Contact List deleted
				log.Printf("Deleted Outbound Contact List %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("error deleting Outbound Contact List %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("Outbound Contact List %s still exists", d.Id()))
	})
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
		contactEmailAddressColumnMap := configEmailColumn.(map[string]interface{})
		if columnName := contactEmailAddressColumnMap["column_name"].(string); columnName != "" {
			sdkContactEmailAddressColumn.ColumnName = &columnName
		}
		if varType := contactEmailAddressColumnMap["type"].(string); varType != "" {
			sdkContactEmailAddressColumn.VarType = &varType
		}
		if contactableTimeColumn := contactEmailAddressColumnMap["contactable_time_column"].(string); contactableTimeColumn != "" {
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
	if columnDataTypeSpecifications == nil || len(columnDataTypeSpecifications) < 1 {
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
	if columnDataTypeSpecifications == nil || len(columnDataTypeSpecifications) == 0 {
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

// type OutboundContactListInstance struct{
// }

// func (*OutboundContactListInstance) ResourceOutboundContactList() *schema.Resource {
// 	ResourceOutboundContactList() *schema.Resource
// }

func GeneratePhoneColumnsBlock(columnName, columnType, callableTimeColumn string) string {
	return fmt.Sprintf(`
	phone_columns {
		column_name          = "%s"
		type                 = "%s"
		callable_time_column = %s
	}
`, columnName, columnType, callableTimeColumn)
}

func GenerateOutboundContactList(
	resourceId string,
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
resource "genesyscloud_outbound_contact_list" "%s" {
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
`, resourceId, name, divisionId, previewModeColumnName, strings.Join(previewModeAcceptedValues, ", "),
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
