package telephony_providers_edges_phonebasesettings

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
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

func createPhoneBaseSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	phoneMetaBase := util.BuildSdkDomainEntityRef(d, "phone_meta_base_id")
	properties := util.BuildTelephonyProperties(d)

	phoneBase := platformclientv2.Phonebase{
		Name:          &name,
		PhoneMetaBase: phoneMetaBase,
		Properties:    properties,
		Capabilities:  buildSdkCapabilities(d),
	}

	if description != "" {
		phoneBase.Description = &description
	}

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	phoneBaseProxy := getPhoneBaseProxy(sdkConfig)

	log.Printf("Getting phone base settings template for %s", phoneMetaBase)
	phoneBaseSettingTemplate, resp, err := phoneBaseProxy.getPhoneBaseSettingTemplate(ctx, *phoneMetaBase.Id)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get phone base settings template %s error: %s", phoneMetaBase, err), resp)
	}

	phoneBaseSettingTemplateLines := *phoneBaseSettingTemplate.Lines
	if len(phoneBaseSettingTemplateLines) == 0 {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get phone base settings template lines for %s", phoneMetaBase), resp)
	}
	phoneBase.Lines = &[]platformclientv2.Linebase{
		{
			Name:         &name,
			LineMetaBase: phoneBaseSettingTemplateLines[0].LineMetaBase,
		},
	}

	log.Printf("Creating phone base settings %s for %s", name, phoneMetaBase)
	phoneBaseSettings, resp, err := phoneBaseProxy.postPhoneBaseSetting(ctx, phoneBase)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create phone base settings %s error: %s", name, err), resp)
	}

	d.SetId(*phoneBaseSettings.Id)

	log.Printf("Created phone base settings %s", *phoneBaseSettings.Id)

	return readPhoneBaseSettings(ctx, d, meta)
}

func updatePhoneBaseSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	phoneMetaBase := util.BuildSdkDomainEntityRef(d, "phone_meta_base_id")
	properties := util.BuildTelephonyProperties(d)
	id := d.Id()

	phoneBase := platformclientv2.Phonebase{
		Id:            &id,
		Name:          &name,
		PhoneMetaBase: phoneMetaBase,
		Properties:    properties,
		Capabilities:  buildSdkCapabilities(d),
	}

	if description != "" {
		phoneBase.Description = &description
	}

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	phoneBaseProxy := getPhoneBaseProxy(sdkConfig)
	phoneBaseSettings, resp, getErr := phoneBaseProxy.getPhoneBaseSetting(ctx, d.Id())
	if getErr != nil {
		if util.IsStatus404(resp) {
			return nil
		}
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read phone base settings %s | error: %s", d.Id(), getErr), resp)
	}

	log.Printf("Getting phone base settings template for %s", phoneMetaBase)
	phoneBaseSettingTemplate, resp, err := phoneBaseProxy.getPhoneBaseSettingTemplate(ctx, *phoneMetaBase.Id)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get phone base settings template %s error: %s", phoneMetaBase, err), resp)
	}

	phoneBaseSettingTemplateLines := *phoneBaseSettingTemplate.Lines
	if len(phoneBaseSettingTemplateLines) == 0 {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get phone base settings template lines for %s", phoneMetaBase), resp)
	}
	phoneBase.Lines = &[]platformclientv2.Linebase{
		{
			Name:         &name,
			LineMetaBase: phoneBaseSettingTemplateLines[0].LineMetaBase,
			Id:           (*phoneBaseSettings.Lines)[0].Id,
			State:        (*phoneBaseSettings.Lines)[0].State,
		},
	}

	log.Printf("Updating phone base settings %s", name)
	_, resp, err = phoneBaseProxy.putPhoneBaseSetting(ctx, d.Id(), phoneBase)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update phone base settings %s error: %s", name, err), resp)
	}

	log.Printf("Updated phone base settings %s", d.Id())

	return readPhoneBaseSettings(ctx, d, meta)
}

func readPhoneBaseSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	phoneBaseProxy := getPhoneBaseProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourcePhoneBaseSettings(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading phone base settings %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		phoneBaseSettings, resp, getErr := phoneBaseProxy.getPhoneBaseSetting(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read phone base settings %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read phone base settings %s | error: %s", d.Id(), getErr), resp))
		}

		d.Set("name", *phoneBaseSettings.Name)

		resourcedata.SetNillableValue(d, "description", phoneBaseSettings.Description)

		if phoneBaseSettings.PhoneMetaBase != nil {
			d.Set("phone_meta_base_id", *phoneBaseSettings.PhoneMetaBase.Id)
		}

		d.Set("properties", nil)
		if phoneBaseSettings.Properties != nil {
			properties, err := util.FlattenTelephonyProperties(phoneBaseSettings.Properties)
			if err != nil {
				return retry.NonRetryableError(fmt.Errorf("%v", err))
			}
			d.Set("properties", properties)
		}

		if phoneBaseSettings.Capabilities != nil {
			d.Set("capabilities", flattenPhoneCapabilities(phoneBaseSettings.Capabilities))
		}

		if len(*phoneBaseSettings.Lines) > 0 {
			d.Set("line_base_settings_id", (*phoneBaseSettings.Lines)[0].Id)
		}

		log.Printf("Read phone base settings %s %s", d.Id(), *phoneBaseSettings.Name)

		return cc.CheckState(d)
	})
}

func deletePhoneBaseSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	phoneBaseProxy := getPhoneBaseProxy(sdkConfig)

	log.Printf("Deleting phone base settings")
	resp, err := phoneBaseProxy.deletePhoneBaseSetting(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete phone base settings %s error: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		phoneBaseSettings, resp, err := phoneBaseProxy.getPhoneBaseSetting(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				// Phone base proxy settings deleted
				log.Printf("Deleted Phone base settings %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error deleting Phone base settings %s | error: %s", d.Id(), err), resp))
		}

		if phoneBaseSettings.State != nil && *phoneBaseSettings.State == "deleted" {
			// Phone base proxy settings deleted
			log.Printf("Deleted Phone base settings %s", d.Id())
			return nil
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("phone base settings %s still exists", d.Id()), resp))
	})
}

func getAllPhoneBaseSettings(ctx context.Context, sdkConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	phoneBaseProxy := getPhoneBaseProxy(sdkConfig)
	phoneBaseSettings, resp, err := phoneBaseProxy.getAllPhoneBaseSettings(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get all phone base settings error: %s", err), resp)
	}

	if phoneBaseSettings != nil {
		for _, phoneBaseSetting := range *phoneBaseSettings {
			resources[*phoneBaseSetting.Id] = &resourceExporter.ResourceMeta{BlockLabel: *phoneBaseSetting.Name}
		}
	}
	return resources, nil
}
