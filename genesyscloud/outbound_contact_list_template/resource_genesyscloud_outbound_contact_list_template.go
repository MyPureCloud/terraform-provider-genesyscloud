package outbound_contact_list_template

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

func getAllOutboundContactListTemplates(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getOutboundContactlisttemplateProxy(clientConfig)

	contactListTemplates, resp, getErr := proxy.getAllOutboundContactlisttemplate(ctx)
	if getErr != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get contact list templates error: %s", getErr), resp)
	}

	for _, contactListTemplate := range *contactListTemplates {
		resources[*contactListTemplate.Id] = &resourceExporter.ResourceMeta{BlockLabel: *contactListTemplate.Name}
	}

	return resources, nil
}

func createOutboundContactListTemplate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	columnNames := lists.InterfaceListToStrings(d.Get("column_names").([]interface{}))
	previewModeColumnName := d.Get("preview_mode_column_name").(string)
	previewModeAcceptedValues := lists.InterfaceListToStrings(d.Get("preview_mode_accepted_values").([]interface{}))
	automaticTimeZoneMapping := d.Get("automatic_time_zone_mapping").(bool)
	zipCodeColumnName := d.Get("zip_code_column_name").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundContactlisttemplateProxy(sdkConfig)

	sdkContactListTemplate := platformclientv2.Contactlisttemplate{
		ColumnNames:                  &columnNames,
		PhoneColumns:                 buildSdkOutboundContactListTemplateContactPhoneNumberColumnSlice(d.Get("phone_columns").(*schema.Set)),
		EmailColumns:                 buildSdkOutboundContactListContactEmailAddressColumnSlice(d.Get("email_columns").(*schema.Set)),
		PreviewModeAcceptedValues:    &previewModeAcceptedValues,
		AttemptLimits:                util.BuildSdkDomainEntityRef(d, "attempt_limit_id"),
		AutomaticTimeZoneMapping:     &automaticTimeZoneMapping,
		ColumnDataTypeSpecifications: buildSdkOutboundContactListTemplateColumnDataTypeSpecifications(d.Get("column_data_type_specifications").([]interface{})),
	}

	if name != "" {
		sdkContactListTemplate.Name = &name
	}
	if previewModeColumnName != "" {
		sdkContactListTemplate.PreviewModeColumnName = &previewModeColumnName
	}
	if zipCodeColumnName != "" {
		sdkContactListTemplate.ZipCodeColumnName = &zipCodeColumnName
	}

	log.Printf("Creating Outbound Contact List Template %s", name)
	outboundContactListTemplate, resp, err := proxy.createOutboundContactlisttemplate(ctx, &sdkContactListTemplate)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create Outbound Contact List Template %s error: %s", name, err), resp)
	}

	d.SetId(*outboundContactListTemplate.Id)

	log.Printf("Created Outbound Contact List Template %s %s", name, *outboundContactListTemplate.Id)
	return readOutboundContactListTemplate(ctx, d, meta)
}

func updateOutboundContactListTemplate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	columnNames := lists.InterfaceListToStrings(d.Get("column_names").([]interface{}))
	previewModeColumnName := d.Get("preview_mode_column_name").(string)
	previewModeAcceptedValues := lists.InterfaceListToStrings(d.Get("preview_mode_accepted_values").([]interface{}))
	automaticTimeZoneMapping := d.Get("automatic_time_zone_mapping").(bool)
	zipCodeColumnName := d.Get("zip_code_column_name").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundContactlisttemplateProxy(sdkConfig)

	sdkContactListTemplate := platformclientv2.Contactlisttemplate{
		ColumnNames:                  &columnNames,
		PhoneColumns:                 buildSdkOutboundContactListTemplateContactPhoneNumberColumnSlice(d.Get("phone_columns").(*schema.Set)),
		EmailColumns:                 buildSdkOutboundContactListContactEmailAddressColumnSlice(d.Get("email_columns").(*schema.Set)),
		PreviewModeAcceptedValues:    &previewModeAcceptedValues,
		AttemptLimits:                util.BuildSdkDomainEntityRef(d, "attempt_limit_id"),
		AutomaticTimeZoneMapping:     &automaticTimeZoneMapping,
		ColumnDataTypeSpecifications: buildSdkOutboundContactListTemplateColumnDataTypeSpecifications(d.Get("column_data_type_specifications").([]interface{})),
	}

	if name != "" {
		sdkContactListTemplate.Name = &name
	}
	if previewModeColumnName != "" {
		sdkContactListTemplate.PreviewModeColumnName = &previewModeColumnName
	}
	if zipCodeColumnName != "" {
		sdkContactListTemplate.ZipCodeColumnName = &zipCodeColumnName
	}

	log.Printf("Updating Outbound Contact List Template %s", name)
	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {

		_, resp, updateErr := proxy.updateOutboundContactlisttemplate(ctx, d.Id(), &sdkContactListTemplate)
		if updateErr != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update Outbound Contact List Template %s error: %s", name, updateErr), resp)
		}
		return nil, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated Outbound Contact List Template %s", name)
	return readOutboundContactListTemplate(ctx, d, meta)
}

func readOutboundContactListTemplate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundContactlisttemplateProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceOutboundContactListTemplate(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading Outbound Contact List Template %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		sdkContactListTemplate, resp, getErr := proxy.getOutboundContactlisttemplateById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read Outbound Contact List Template %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read Outbound Contact List Template %s | error: %s", d.Id(), getErr), resp))
		}

		if sdkContactListTemplate.Name != nil {
			_ = d.Set("name", *sdkContactListTemplate.Name)
		}
		if sdkContactListTemplate.ColumnNames != nil {
			var columnNames []string
			for _, name := range *sdkContactListTemplate.ColumnNames {
				columnNames = append(columnNames, name)
			}
			_ = d.Set("column_names", columnNames)
		}
		if sdkContactListTemplate.PhoneColumns != nil {
			_ = d.Set("phone_columns", flattenSdkOutboundContactListTemplateContactPhoneNumberColumnSlice(*sdkContactListTemplate.PhoneColumns))
		}
		if sdkContactListTemplate.EmailColumns != nil {
			_ = d.Set("email_columns", flattenSdkOutboundContactListTemplateContactEmailAddressColumnSlice(*sdkContactListTemplate.EmailColumns))
		}
		if sdkContactListTemplate.PreviewModeColumnName != nil {
			_ = d.Set("preview_mode_column_name", *sdkContactListTemplate.PreviewModeColumnName)
		}
		if sdkContactListTemplate.PreviewModeAcceptedValues != nil {
			var acceptedValues []string
			for _, val := range *sdkContactListTemplate.PreviewModeAcceptedValues {
				acceptedValues = append(acceptedValues, val)
			}
			_ = d.Set("preview_mode_accepted_values", acceptedValues)
		}
		if sdkContactListTemplate.AttemptLimits != nil && sdkContactListTemplate.AttemptLimits.Id != nil {
			_ = d.Set("attempt_limit_id", *sdkContactListTemplate.AttemptLimits.Id)
		}
		if sdkContactListTemplate.AutomaticTimeZoneMapping != nil {
			_ = d.Set("automatic_time_zone_mapping", *sdkContactListTemplate.AutomaticTimeZoneMapping)
		}
		if sdkContactListTemplate.ZipCodeColumnName != nil {
			_ = d.Set("zip_code_column_name", *sdkContactListTemplate.ZipCodeColumnName)
		}
		if sdkContactListTemplate.ColumnDataTypeSpecifications != nil {
			_ = d.Set("column_data_type_specifications", flattenSdkOutboundContactListTemplateColumnDataTypeSpecifications(*sdkContactListTemplate.ColumnDataTypeSpecifications))
		}

		log.Printf("Read Outbound Contact List Template %s %s", d.Id(), *sdkContactListTemplate.Name)
		return cc.CheckState(d)
	})
}

func deleteOutboundContactListTemplate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundContactlisttemplateProxy(sdkConfig)

	diagErr := util.RetryWhen(util.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting Outbound Contact List Template")
		resp, err := proxy.deleteOutboundContactlisttemplate(ctx, d.Id())
		if err != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete Outbound Contact List Template %s error: %s", d.Id(), err), resp)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getOutboundContactlisttemplateById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				// Outbound Contact List Template deleted
				log.Printf("Deleted Outbound Contact List Template %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error deleting Outbound Contact List Template %s | error: %s", d.Id(), err), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Outbound Contact List Template %s still exists", d.Id()), resp))
	})
}
