package genesyscloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v53/platformclientv2"
	"log"
)

var (
	propertiesPhone_label             = "phone_label"
	propertiesPhone_mwi_enabled       = "phone_mwi_enabled"
	propertiesPhone_mwi_subscribe     = "phone_mwi_subscribe"
	propertiesPhone_stations          = "phone_stations"
	propertiesPhone_maxLineKeys       = "phone_max_line_keys"
	propertiesPhone_standalone        = "phone_standalone"
	propertiesPhone_ignoreOnSecondary = "phone_ignore_on_secondary"

	propertiesPhone_media_codecs = "phone_media_codecs"
	propertiesPhone_media_dscp   = "phone_media_dscp"
	propertiesPhone_sip_dscp     = "phone_sip_dscp"

	phoneBaseSettingsProperties = &schema.Resource{
		Schema: map[string]*schema.Schema{
			propertiesPhone_label: {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "WebRTC Phone Connections",
			},
			propertiesPhone_mwi_enabled: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			propertiesPhone_mwi_subscribe: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			propertiesPhone_maxLineKeys: {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
			},
			propertiesPhone_stations: {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type:    schema.TypeString,
					Default: nil,
				},
			},
			propertiesPhone_standalone: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			propertiesPhone_ignoreOnSecondary: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			propertiesPhone_media_codecs: {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type:    schema.TypeString,
					Default: "audio/opus",
				},
			},
			propertiesPhone_media_dscp: {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      46,
				ValidateFunc: validation.IntBetween(0, 63),
			},
			propertiesPhone_sip_dscp: {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      24,
				ValidateFunc: validation.IntBetween(0, 63),
			},
		},
	}
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
				Description: "phone base settings properties",
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem:        phoneBaseSettingsProperties,
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
	}
}

func buildSdkPhoneBaseSettingsProperty(property interface{}) map[string]interface{} {
	return buildSdkTrunkBaseSettingsProperty(property)
}

func buildSdkPhoneBaseSettingsProperties(d *schema.ResourceData) *map[string]interface{} {
	returnValue := make(map[string]interface{})

	if properties := d.Get("properties"); properties != nil {
		prop := properties.(*schema.Set)
		propList := prop.List()

		if len(propList) == 0 {
			return &returnValue
		}

		propertiesMap := propList[0].(map[string]interface{})
		if property, ok := propertiesMap[propertiesPhone_label].(string); ok && property != "" {
			returnValue["phone_label"] = buildSdkPhoneBaseSettingsProperty(property)
		}
		if property, ok := propertiesMap[propertiesPhone_mwi_enabled].(bool); ok {
			returnValue["phone_mwi_enabled"] = buildSdkPhoneBaseSettingsProperty(property)
		}
		if property, ok := propertiesMap[propertiesPhone_mwi_subscribe].(bool); ok {
			returnValue["phone_mwi_subscribe"] = buildSdkPhoneBaseSettingsProperty(property)
		}
		if property, ok := propertiesMap[propertiesPhone_stations].([]interface{}); ok {
			if len(property) == 0 {
				returnValue["phone_stations"] = buildSdkPhoneBaseSettingsProperty(nil)
			} else {
				returnValue["phone_stations"] = buildSdkPhoneBaseSettingsProperty(property)
			}
		}
		if property, ok := propertiesMap[propertiesPhone_maxLineKeys].(int); ok {
			returnValue["phone_maxLineKeys"] = buildSdkPhoneBaseSettingsProperty(property)
		}
		if property, ok := propertiesMap[propertiesPhone_standalone].(bool); ok {
			returnValue["phone_standalone"] = buildSdkPhoneBaseSettingsProperty(property)
		}
		if property, ok := propertiesMap[propertiesPhone_ignoreOnSecondary].(bool); ok {
			returnValue["phone_ignoreOnSecondary"] = buildSdkPhoneBaseSettingsProperty(property)
		}
		if property, ok := propertiesMap[propertiesPhone_media_codecs].([]interface{}); ok {
			if len(property) == 0 {
				returnValue["phone_media_codecs"] = buildSdkPhoneBaseSettingsProperty(nil)
			} else {
				returnValue["phone_media_codecs"] = buildSdkPhoneBaseSettingsProperty(property)
			}
		}
		if property, ok := propertiesMap[propertiesPhone_media_dscp].(int); ok {
			returnValue["phone_media_dscp"] = buildSdkPhoneBaseSettingsProperty(property)
		}
		if property, ok := propertiesMap[propertiesPhone_sip_dscp].(int); ok {
			returnValue["phone_sip_dscp"] = buildSdkPhoneBaseSettingsProperty(property)
		}
	}

	return &returnValue
}

func createPhoneBaseSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	phoneMetaBase := buildSdkDomainEntityRef(d, "phone_meta_base_id")
	properties := buildSdkPhoneBaseSettingsProperties(d)

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

	sdkConfig := meta.(*providerMeta).ClientConfig
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
	properties := buildSdkPhoneBaseSettingsProperties(d)
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

	sdkConfig := meta.(*providerMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	// Get the latest version of the setting
	phoneBaseSettings, resp, getErr := edgesAPI.GetTelephonyProvidersEdgesPhonebasesetting(d.Id())
	if getErr != nil {
		if resp != nil && resp.StatusCode == 404 {
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
	sdkConfig := meta.(*providerMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	log.Printf("Reading phone base settings %s", d.Id())
	phoneBaseSettings, resp, getErr := edgesAPI.GetTelephonyProvidersEdgesPhonebasesetting(d.Id())

	if getErr != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read phone base settings %s: %s", d.Id(), getErr)
	}

	d.Set("name", *phoneBaseSettings.Name)
	if phoneBaseSettings.Description != nil {
		d.Set("description", *phoneBaseSettings.Description)
	}
	if phoneBaseSettings.PhoneMetaBase != nil {
		d.Set("phone_meta_base_id", *phoneBaseSettings.PhoneMetaBase.Id)
	}

	d.Set("properties", nil)
	if phoneBaseSettings.Properties != nil {
		d.Set("properties", flattenPhoneBaseSettingsProperties(phoneBaseSettings.Properties))
	}

	if phoneBaseSettings.Capabilities != nil {
		d.Set("capabilities", flattenPhoneCapabilities(phoneBaseSettings.Capabilities))
	}

	if len(*phoneBaseSettings.Lines) > 0 {
		d.Set("line_base_settings_id", (*phoneBaseSettings.Lines)[0].Id)
	}

	log.Printf("Read phone base settings %s %s", d.Id(), *phoneBaseSettings.Name)

	return nil
}

func flattenPhoneBaseSettingsProperty(property map[string]interface{}) interface{} {
	return flattenTrunkBaseSettingsProperty(property)
}

func flattenPhoneBaseSettingsProperties(properties interface{}) *schema.Set {
	propertyMap := make(map[string]interface{})

	propertyV := *(properties.(*map[string]interface{}))

	if property, ok := propertyV["phone_label"].(map[string]interface{}); ok {
		if flattenedProperty := flattenPhoneBaseSettingsProperty(property); flattenedProperty != nil {
			propertyMap[propertiesPhone_label] = flattenedProperty.(string)
		}
	}
	if property, ok := propertyV["phone_mwi_enabled"].(map[string]interface{}); ok {
		if flattenedProperty := flattenPhoneBaseSettingsProperty(property); flattenedProperty != nil {
			propertyMap[propertiesPhone_mwi_enabled] = flattenedProperty.(bool)
		}
	}
	if property, ok := propertyV["phone_mwi_subscribe"].(map[string]interface{}); ok {
		if flattenedProperty := flattenPhoneBaseSettingsProperty(property); flattenedProperty != nil {
			propertyMap[propertiesPhone_mwi_subscribe] = flattenedProperty.(bool)
		}
	}
	if property, ok := propertyV["phone_maxLineKeys"].(map[string]interface{}); ok {
		if flattenedProperty := flattenPhoneBaseSettingsProperty(property); flattenedProperty != nil {
			propertyMap[propertiesPhone_maxLineKeys] = int(flattenedProperty.(float64))
		}
	}
	if property, ok := propertyV["phone_stations"].(map[string]interface{}); ok {
		if flattenedProperty := flattenPhoneBaseSettingsProperty(property); flattenedProperty != nil {
			flattenedPropertySlice := flattenedProperty.([]interface{})
			propertySlice := make([]interface{}, 0)
			for _, flattenedPropertyValue := range flattenedPropertySlice {
				propertySlice = append(propertySlice, flattenedPropertyValue.(interface{}))
			}
			propertyMap[propertiesPhone_stations] = propertySlice
		}
	}
	if property, ok := propertyV["phone_standalone"].(map[string]interface{}); ok {
		if flattenedProperty := flattenPhoneBaseSettingsProperty(property); flattenedProperty != nil {
			propertyMap[propertiesPhone_standalone] = flattenedProperty.(bool)
		}
	}
	if property, ok := propertyV["phone_ignoreOnSecondary"].(map[string]interface{}); ok {
		if flattenedProperty := flattenPhoneBaseSettingsProperty(property); flattenedProperty != nil {
			propertyMap[propertiesPhone_ignoreOnSecondary] = flattenedProperty.(bool)
		}
	}
	if property, ok := propertyV["phone_media_codecs"].(map[string]interface{}); ok {
		if flattenedProperty := flattenPhoneBaseSettingsProperty(property); flattenedProperty != nil {
			flattenedPropertySlice := flattenedProperty.([]interface{})
			propertySlice := make([]interface{}, 0)
			for _, flattenedPropertyValue := range flattenedPropertySlice {
				propertySlice = append(propertySlice, flattenedPropertyValue.(interface{}))
			}
			propertyMap[propertiesPhone_media_codecs] = propertySlice
		}
	}
	if property, ok := propertyV["phone_media_dscp"].(map[string]interface{}); ok {
		if flattenedProperty := flattenPhoneBaseSettingsProperty(property); flattenedProperty != nil {
			propertyMap[propertiesPhone_media_dscp] = int(flattenedProperty.(float64))
		}
	}
	if property, ok := propertyV["phone_sip_dscp"].(map[string]interface{}); ok {
		if flattenedProperty := flattenPhoneBaseSettingsProperty(property); flattenedProperty != nil {
			propertyMap[propertiesPhone_sip_dscp] = int(flattenedProperty.(float64))
		}
	}

	propertySet := schema.NewSet(schema.HashResource(phoneBaseSettingsProperties), []interface{}{})
	propertySet.Add(propertyMap)

	return propertySet
}

func deletePhoneBaseSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	log.Printf("Deleting phone base settings")
	_, err := edgesAPI.DeleteTelephonyProvidersEdgesPhonebasesetting(d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete phone base settings: %s", err)
	}
	log.Printf("Deleted phone base settings")
	return nil
}

func getAllPhoneBaseSettings(ctx context.Context, sdkConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(ResourceIDMetaMap)

	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	for pageNum := 1; ; pageNum++ {
		phoneBaseSettings, _, getErr := edgesAPI.GetTelephonyProvidersEdgesPhonebasesettings(pageNum, 100, "", "", nil, "")
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of phone base settings: %v", getErr)
		}

		if phoneBaseSettings.Entities == nil || len(*phoneBaseSettings.Entities) == 0 {
			break
		}

		for _, phoneBaseSetting := range *phoneBaseSettings.Entities {
			if *phoneBaseSetting.State != "deleted" {
				resources[*phoneBaseSetting.Id] = &ResourceMeta{Name: *phoneBaseSetting.Name}
			}
		}
	}

	return resources, nil
}

func phoneBaseSettingsExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllPhoneBaseSettings),
		RefAttrs:         map[string]*RefAttrSettings{},
	}
}
