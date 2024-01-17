package telephony_providers_edges_phonebasesettings

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"
)

func createPhoneBaseSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	phoneMetaBase := gcloud.BuildSdkDomainEntityRef(d, "phone_meta_base_id")
	properties := gcloud.BuildBaseSettingsProperties(d)

	phoneBase := platformclientv2.Phonebase{
		Name:          &name,
		PhoneMetaBase: phoneMetaBase,
		Properties:    properties,
		Lines: &[]platformclientv2.Linebase{
			{
				Name:         &name,
				LineMetaBase: phoneMetaBase,
			},
		},
		Capabilities: buildSdkCapabilities(d),
	}

	if description != "" {
		phoneBase.Description = &description
	}

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	phoneBaseProxy := getPhoneBaseProxy(sdkConfig)

	log.Printf("Creating phone base settings %s", name)
	phoneBaseSettings, _, err := phoneBaseProxy.postPhoneBaseSetting(ctx, phoneBase)
	if err != nil {
		return diag.Errorf("Failed to create phone base settings %s: %s", name, err)
	}

	d.SetId(*phoneBaseSettings.Id)

	log.Printf("Created phone base settings %s", *phoneBaseSettings.Id)

	return readPhoneBaseSettings(ctx, d, meta)
}

func updatePhoneBaseSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	phoneMetaBase := gcloud.BuildSdkDomainEntityRef(d, "phone_meta_base_id")
	properties := gcloud.BuildBaseSettingsProperties(d)
	id := d.Id()

	phoneBase := platformclientv2.Phonebase{
		Id:            &id,
		Name:          &name,
		PhoneMetaBase: phoneMetaBase,
		Properties:    properties,
		Lines: &[]platformclientv2.Linebase{
			{
				Name:         &name,
				LineMetaBase: phoneMetaBase,
			},
		},
		Capabilities: buildSdkCapabilities(d),
	}

	if description != "" {
		phoneBase.Description = &description
	}

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	phoneBaseProxy := getPhoneBaseProxy(sdkConfig)
	phoneBaseSettings, resp, getErr := phoneBaseProxy.getPhoneBaseSetting(ctx, d.Id())
	if getErr != nil {
		if gcloud.IsStatus404(resp) {
			return nil
		}
		return diag.Errorf("Failed to read phone base settings %s: %s", d.Id(), getErr)
	}
	(*phoneBase.Lines)[0].Id = (*phoneBaseSettings.Lines)[0].Id

	log.Printf("Updating phone base settings %s", name)
	phoneBaseSettings, resp, err := phoneBaseProxy.putPhoneBaseSetting(ctx, d.Id(), phoneBase)
	if err != nil {
		return diag.Errorf("Failed to update phone base settings %s: %v", name, err)
	}

	log.Printf("Updated phone base settings %s", d.Id())

	return readPhoneBaseSettings(ctx, d, meta)
}

func readPhoneBaseSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	phoneBaseProxy := getPhoneBaseProxy(sdkConfig)
	log.Printf("Reading phone base settings %s", d.Id())

	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		phoneBaseSettings, resp, getErr := phoneBaseProxy.getPhoneBaseSetting(ctx, d.Id())
		if getErr != nil {
			if gcloud.IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("failed to read phone base settings %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("failed to read phone base settings %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourcePhoneBaseSettings())
		d.Set("name", *phoneBaseSettings.Name)

		resourcedata.SetNillableValue(d, "description", phoneBaseSettings.Description)

		if phoneBaseSettings.PhoneMetaBase != nil {
			d.Set("phone_meta_base_id", *phoneBaseSettings.PhoneMetaBase.Id)
		}

		d.Set("properties", nil)
		if phoneBaseSettings.Properties != nil {
			properties, err := gcloud.FlattenBaseSettingsProperties(phoneBaseSettings.Properties)
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

		return cc.CheckState()
	})
}

func deletePhoneBaseSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	phoneBaseProxy := getPhoneBaseProxy(sdkConfig)

	log.Printf("Deleting phone base settings")
	_, err := phoneBaseProxy.deletePhoneBaseSetting(ctx, d.Id())
	if err != nil {
		return diag.Errorf("failed to delete phone base settings: %s", err)
	}

	return gcloud.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		phoneBaseSettings, resp, err := phoneBaseProxy.getPhoneBaseSetting(ctx, d.Id())
		if err != nil {
			if gcloud.IsStatus404(resp) {
				// Phone base proxy settings deleted
				log.Printf("Deleted Phone base settings %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("error deleting Phone base settings %s: %s", d.Id(), err))
		}

		if phoneBaseSettings.State != nil && *phoneBaseSettings.State == "deleted" {
			// Phone base proxy settings deleted
			log.Printf("Deleted Phone base settings %s", d.Id())
			return nil
		}

		return retry.RetryableError(fmt.Errorf("phone base settings %s still exists", d.Id()))
	})
}

func getAllPhoneBaseSettings(ctx context.Context, sdkConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	phoneBaseProxy := getPhoneBaseProxy(sdkConfig)
	phoneBaseSettings, err := phoneBaseProxy.getAllPhoneBaseSettings(ctx)
	if err != nil {
		return nil, diag.Errorf("failed to get all phone base settings: %s", err)
	}

	if phoneBaseSettings != nil {
		for _, phoneBaseSetting := range *phoneBaseSettings {
			resources[*phoneBaseSetting.Id] = &resourceExporter.ResourceMeta{Name: *phoneBaseSetting.Name}
		}
	}
	return resources, nil
}
