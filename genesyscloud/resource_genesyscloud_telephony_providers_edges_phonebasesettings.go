package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"time"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v99/platformclientv2"
)

func resourcePhoneBaseSettings() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Phone Base Settings",

		CreateContext: createWithPooledClient(createPhoneBaseSettings),
		ReadContext:   readWithPooledClient(readPhoneBaseSettings),
		UpdateContext: updateWithPooledClient(updatePhoneBaseSettings),
		DeleteContext: deleteWithPooledClient(deletePhoneBaseSettings),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the entity.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "The resource's description.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"phone_meta_base_id": {
				Description: "A phone metabase is essentially a database for storing phone configuration settings, which simplifies the configuration process.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"properties": {
				Description:      "phone base settings properties",
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: suppressEquivalentJsonDiffs,
			},
			"capabilities": {
				Description: "Phone Capabilities.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem:        phoneCapabilities,
			},
			"line_base_settings_id": {
				Description: "Computed line base settings id",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
		},
		CustomizeDiff: customizePhoneBaseSettingsPropertiesDiff,
	}
}

func createPhoneBaseSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	phoneMetaBase := buildSdkDomainEntityRef(d, "phone_meta_base_id")
	properties := buildBaseSettingsProperties(d)

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

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	log.Printf("Creating phone base settings %s", name)
	phoneBaseSettings, _, err := edgesAPI.PostTelephonyProvidersEdgesPhonebasesettings(phoneBase)
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
	phoneMetaBase := buildSdkDomainEntityRef(d, "phone_meta_base_id")
	properties := buildBaseSettingsProperties(d)
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

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	phoneBaseSettings, resp, getErr := edgesAPI.GetTelephonyProvidersEdgesPhonebasesetting(d.Id())
	if getErr != nil {
		if isStatus404(resp) {
			return nil
		}
		return diag.Errorf("Failed to read phone base settings %s: %s", d.Id(), getErr)
	}
	(*phoneBase.Lines)[0].Id = (*phoneBaseSettings.Lines)[0].Id

	log.Printf("Updating phone base settings %s", name)
	phoneBaseSettings, resp, err := edgesAPI.PutTelephonyProvidersEdgesPhonebasesetting(d.Id(), phoneBase)
	if err != nil {
		return diag.Errorf("Failed to update phone base settings %s: %v", name, err)
	}

	log.Printf("Updated phone base settings %s", d.Id())

	return readPhoneBaseSettings(ctx, d, meta)
}

func readPhoneBaseSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	log.Printf("Reading phone base settings %s", d.Id())
	return withRetriesForRead(ctx, d, func() *resource.RetryError {
		phoneBaseSettings, resp, getErr := edgesAPI.GetTelephonyProvidersEdgesPhonebasesetting(d.Id())
		if getErr != nil {
			if isStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to read phone base settings %s: %s", d.Id(), getErr))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to read phone base settings %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, resourcePhoneBaseSettings())
		d.Set("name", *phoneBaseSettings.Name)
		if phoneBaseSettings.Description != nil {
			d.Set("description", *phoneBaseSettings.Description)
		}
		if phoneBaseSettings.PhoneMetaBase != nil {
			d.Set("phone_meta_base_id", *phoneBaseSettings.PhoneMetaBase.Id)
		}

		d.Set("properties", nil)
		if phoneBaseSettings.Properties != nil {
			properties, err := flattenBaseSettingsProperties(phoneBaseSettings.Properties)
			if err != nil {
				return resource.NonRetryableError(fmt.Errorf("%v", err))
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
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	log.Printf("Deleting phone base settings")
	_, err := edgesAPI.DeleteTelephonyProvidersEdgesPhonebasesetting(d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete phone base settings: %s", err)
	}

	return withRetries(ctx, 30*time.Second, func() *resource.RetryError {
		phoneBaseSettings, resp, err := edgesAPI.GetTelephonyProvidersEdgesPhonebasesetting(d.Id())
		if err != nil {
			if isStatus404(resp) {
				// Phone base settings deleted
				log.Printf("Deleted Phone base settings %s", d.Id())
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("Error deleting Phone base settings %s: %s", d.Id(), err))
		}

		if phoneBaseSettings.State != nil && *phoneBaseSettings.State == "deleted" {
			// Phone base settings deleted
			log.Printf("Deleted Phone base settings %s", d.Id())
			return nil
		}

		return resource.RetryableError(fmt.Errorf("Phone base settings %s still exists", d.Id()))
	})
}

func getAllPhoneBaseSettings(ctx context.Context, sdkConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(ResourceIDMetaMap)

	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	err := withRetries(ctx, 15*time.Second, func() *resource.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			phoneBaseSettings, resp, getErr := edgesAPI.GetTelephonyProvidersEdgesPhonebasesettings(pageSize, pageNum, "", "", nil, "")
			if getErr != nil {
				if isStatus404(resp) {
					return resource.RetryableError(fmt.Errorf("Failed to get page of phonebasesettings: %v", getErr))
				}
				return resource.NonRetryableError(fmt.Errorf("Failed to get page of phonebasesettings: %v", getErr))
			}

			if phoneBaseSettings.Entities == nil || len(*phoneBaseSettings.Entities) == 0 {
				break
			}

			for _, phoneBaseSetting := range *phoneBaseSettings.Entities {
				if phoneBaseSetting.State != nil && *phoneBaseSetting.State != "deleted" {
					resources[*phoneBaseSetting.Id] = &ResourceMeta{Name: *phoneBaseSetting.Name}
				}
			}
		}

		return nil
	})

	return resources, err
}

func phoneBaseSettingsExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc:     getAllWithPooledClient(getAllPhoneBaseSettings),
		RefAttrs:             map[string]*RefAttrSettings{},
		JsonEncodeAttributes: []string{"properties"},
	}
}
