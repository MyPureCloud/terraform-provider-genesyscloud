package genesyscloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v48/platformclientv2"
	"log"
	"regexp"
)

var (
	propertiesTrunk_type                                = "trunk_type"
	propertiesTrunk_label                               = "trunk_label"
	propertiesTrunk_enabled                             = "trunk_enabled"
	propertiesTrunk_maxDialTimeout                      = "trunk_max_dial_timeout"
	propertiesTrunk_maxCallRate                         = "trunk_max_call_rate"
	propertiesTrunk_transport_sipDscpValue              = "trunk_transport_sip_dscp_value"
	propertiesTrunk_transport_tcp_connectTimeout        = "trunk_transport_tcp_connect_timeout"
	propertiesTrunk_transport_tcp_connectionIdleTimeout = "trunk_transport_tcp_connection_idle_timeout"
	propertiesTrunk_transport_retryableReasonCodes      = "trunk_transport_retryable_reason_codes"
	propertiesTrunk_transport_retryableCauseCodes       = "trunk_transport_retryable_cause_codes"
	propertiesTrunk_media_codec                         = "trunk_media_codec"
	propertiesTrunk_media_dtmf_method                   = "trunk_media_dtmf_method"
	propertiesTrunk_media_dtmf_payload                  = "trunk_media_dtmf_payload"
	propertiesTrunk_media_dscpValue                     = "trunk_media_dscp_value"
	propertiesTrunk_media_srtpCipherSuites              = "trunk_media_srtp_cipher_suites"
	propertiesTrunk_media_disconnectOnIdleRTP           = "trunk_media_disconnect_on_idle_rtp"
	propertiesTrunk_diagnostic_capture_enabled          = "trunk_diagnostic_capture_enabled"
	propertiesTrunk_diagnostic_capture_endTime          = "trunk_diagnostic_capture_end_time"
	propertiesTrunk_diagnostic_protocol_endTime         = "trunk_diagnostic_protocol_end_time"
	propertiesTrunk_language                            = "trunk_language"

	trunkBaseSettingsProperties = &schema.Resource{
		Schema: map[string]*schema.Schema{
			propertiesTrunk_type: {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "station",
				ValidateFunc: validation.StringInSlice([]string{
					"external",
					"tie",
					"tie.direct",
					"tie.indirect",
					"tie.cloud.proxy",
					"station",
					"station.cdm",
					"station.cdm.webrtc",
					"external.pcv",
					"external.pcv.aws",
					"external.pcv.byoc.carrier",
					"external.pcv.byoc.pbx"}, false),
			},
			propertiesTrunk_label: {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "WebRTC Phone Connections",
			},
			propertiesTrunk_enabled: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			propertiesTrunk_maxDialTimeout: {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "1m",
			},
			propertiesTrunk_maxCallRate: {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "40/5s",
			},
			propertiesTrunk_transport_sipDscpValue: {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      24,
				ValidateFunc: validation.IntBetween(-1, 63),
			},
			propertiesTrunk_transport_tcp_connectTimeout: {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      2,
				ValidateFunc: validation.IntBetween(1, 60),
			},
			propertiesTrunk_transport_tcp_connectionIdleTimeout: {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      86400,
				ValidateFunc: validation.IntBetween(30, 86400),
			},
			propertiesTrunk_transport_retryableReasonCodes: {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "500-599",
			},
			propertiesTrunk_transport_retryableCauseCodes: {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "1-5,25,27,28,31,34,38,41,42,44,46,62,63,79,91,96,97,99,100,103",
			},
			propertiesTrunk_media_codec: {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Default:      "audio/opus",
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{"audio/opus", "audio/pcmu", "audio/pcma", "audio/g729", "audio/g722"}, false),
				},
			},
			propertiesTrunk_media_dtmf_method: {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "RTP Events",
				ValidateFunc: validation.StringInSlice([]string{"None", "RTP Events", "In-band Audio"}, false),
			},
			propertiesTrunk_media_dtmf_payload: {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      101,
				ValidateFunc: validation.IntBetween(96, 127),
			},
			propertiesTrunk_media_dscpValue: {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      46,
				ValidateFunc: validation.IntBetween(-1, 63),
			},
			propertiesTrunk_media_srtpCipherSuites: {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Default: "AES_CM_128_HMAC_SHA1_80",
					Type:    schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						"AES_CM_128_HMAC_SHA1_80",
						"AES_CM_128_HMAC_SHA1_32",
						"AES_CM_192_HMAC_SHA1_80",
						"AES_CM_192_HMAC_SHA1_32",
						"AES_CM_256_HMAC_SHA1_80",
						"AES_CM_256_HMAC_SHA1_32"}, false),
				},
			},
			propertiesTrunk_media_disconnectOnIdleRTP: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			propertiesTrunk_diagnostic_capture_enabled: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			propertiesTrunk_diagnostic_capture_endTime: {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      nil,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^\\d{4}-[0-1][0-9]-[0-3][0-9](T| )[0-2][0-9]:[0-5][0-9]:[0-5][0-9](\\.(\\d{1})*)?(Z| )*$`), ""),
			},
			propertiesTrunk_diagnostic_protocol_endTime: {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      nil,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^\\d{4}-[0-1][0-9]-[0-3][0-9](T| )[0-2][0-9]:[0-5][0-9]:[0-5][0-9](\\.(\\d{1})*)?(Z| )*$`), ""),
			},
			propertiesTrunk_language: {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "en-US",
				ValidateFunc: validation.StringInSlice([]string{
					"en-AU",
					"pt-BR",
					"fr-CA",
					"zh-CN",
					"de-DE",
					"en-GB",
					"it-IT",
					"ja-JP",
					"nl-NL",
					"en-US",
					"es-US",
				}, false),
			},
		},
	}
)

func resourceTrunkBaseSettings() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Trunk Base Settings",

		CreateContext: createWithPooledClient(createTrunkBaseSettings),
		ReadContext:   readWithPooledClient(readTrunkBaseSettings),
		UpdateContext: updateWithPooledClient(updateTrunkBaseSettings),
		DeleteContext: deleteWithPooledClient(deleteTrunkBaseSettings),
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
			"state": {
				Description: "The resource's state.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"description": {
				Description: "The resource's description.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"trunk_meta_base_id": {
				Description: "The meta-base this trunk is based on.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"properties": {
				Description: "trunk base settings properties",
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem:        trunkBaseSettingsProperties,
			},
			"trunk_type": {
				Description:  "The type of this trunk base.Valid values: EXTERNAL, PHONE, EDGE.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"EXTERNAL", "PHONE", "EDGE"}, false),
			},
			"managed": {
				Description: "Is this trunk being managed remotely. This property is synchronized with the managed property of the Edge Group to which it is assigned.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
		},
	}
}

func buildSdkTrunkBaseSettingsProperty(property interface{}) map[string]interface{} {
	return map[string]interface{}{
		"value": map[string]interface{}{
			"instance": property,
		},
	}
}

func buildSdkTrunkBaseSettingsProperties(d *schema.ResourceData) *map[string]interface{} {
	returnValue := make(map[string]interface{})

	if properties := d.Get("properties"); properties != nil {
		prop := properties.(*schema.Set)
		propList := prop.List()

		if len(propList) == 0 {
			return &returnValue
		}

		propertiesMap := propList[0].(map[string]interface{})
		if property, ok := propertiesMap[propertiesTrunk_type].(string); ok && property != "" {
			returnValue["trunk_type"] = buildSdkTrunkBaseSettingsProperty(property)
		}
		if property, ok := propertiesMap[propertiesTrunk_label].(string); ok && property != "" {
			returnValue["trunk_label"] = buildSdkTrunkBaseSettingsProperty(property)
		}
		if property, ok := propertiesMap[propertiesTrunk_enabled].(bool); ok {
			returnValue["trunk_enabled"] = buildSdkTrunkBaseSettingsProperty(property)
		}
		if property, ok := propertiesMap[propertiesTrunk_maxDialTimeout].(string); ok && property != "" {
			returnValue["trunk_maxDialTimeout"] = buildSdkTrunkBaseSettingsProperty(property)
		}
		if property, ok := propertiesMap[propertiesTrunk_maxCallRate].(string); ok && property != "" {
			returnValue["trunk_maxCallRate"] = buildSdkTrunkBaseSettingsProperty(property)
		}
		if property, ok := propertiesMap[propertiesTrunk_transport_sipDscpValue].(int); ok {
			returnValue["trunk_transport_sipDscpValue"] = buildSdkTrunkBaseSettingsProperty(property)
		}
		if property, ok := propertiesMap[propertiesTrunk_transport_tcp_connectTimeout].(int); ok {
			returnValue["trunk_transport_tcp_connectTimeout"] = buildSdkTrunkBaseSettingsProperty(property)
		}
		if property, ok := propertiesMap[propertiesTrunk_transport_tcp_connectionIdleTimeout].(int); ok {
			returnValue["trunk_transport_tcp_connectionIdleTimeout"] = buildSdkTrunkBaseSettingsProperty(property)
		}
		if property, ok := propertiesMap[propertiesTrunk_transport_retryableReasonCodes].(string); ok && property != "" {
			returnValue["trunk_transport_retryableReasonCodes"] = buildSdkTrunkBaseSettingsProperty(property)
		}
		if property, ok := propertiesMap[propertiesTrunk_transport_retryableCauseCodes].(string); ok && property != "" {
			returnValue["trunk_transport_retryableCauseCodes"] = buildSdkTrunkBaseSettingsProperty(property)
		}
		if property, ok := propertiesMap[propertiesTrunk_media_codec].([]interface{}); ok {
			returnValue["trunk_media_codec"] = buildSdkTrunkBaseSettingsProperty(property)
		}
		if property, ok := propertiesMap[propertiesTrunk_media_dtmf_method].(string); ok && property != "" {
			returnValue["trunk_media_dtmf_method"] = buildSdkTrunkBaseSettingsProperty(property)
		}
		if property, ok := propertiesMap[propertiesTrunk_media_dtmf_payload].(int); ok {
			returnValue["trunk_media_dtmf_payload"] = buildSdkTrunkBaseSettingsProperty(property)
		}
		if property, ok := propertiesMap[propertiesTrunk_media_dscpValue].(int); ok {
			returnValue["trunk_media_dscpValue"] = buildSdkTrunkBaseSettingsProperty(property)
		}
		if property, ok := propertiesMap[propertiesTrunk_media_srtpCipherSuites].([]interface{}); ok {
			returnValue["trunk_media_srtpCipherSuites"] = buildSdkTrunkBaseSettingsProperty(property)
		}
		if property, ok := propertiesMap[propertiesTrunk_media_disconnectOnIdleRTP].(bool); ok {
			returnValue["trunk_media_disconnectOnIdleRTP"] = buildSdkTrunkBaseSettingsProperty(property)
		}
		if property, ok := propertiesMap[propertiesTrunk_diagnostic_capture_enabled].(bool); ok {
			returnValue["trunk_diagnostic_capture_enabled"] = buildSdkTrunkBaseSettingsProperty(property)
		}
		if property, ok := propertiesMap[propertiesTrunk_diagnostic_capture_endTime].(string); ok && property != "" {
			returnValue["trunk_diagnostic_capture_endTime"] = buildSdkTrunkBaseSettingsProperty(property)
		}
		if property, ok := propertiesMap[propertiesTrunk_diagnostic_protocol_endTime].(string); ok && property != "" {
			returnValue["trunk_diagnostic_protocol_endTime"] = buildSdkTrunkBaseSettingsProperty(property)
		}
		if property, ok := propertiesMap[propertiesTrunk_language].(string); ok && property != "" {
			returnValue["trunk_language"] = buildSdkTrunkBaseSettingsProperty(property)
		}
	}

	return &returnValue
}

func createTrunkBaseSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	trunkMetaBase := buildSdkDomainEntityRef(d, "trunk_meta_base_id")
	properties := buildSdkTrunkBaseSettingsProperties(d)

	trunkType := d.Get("trunk_type").(string)
	managed := d.Get("managed").(bool)

	trunkBase := platformclientv2.Trunkbase{
		Name:          &name,
		TrunkMetabase: trunkMetaBase,
		TrunkType:     &trunkType,
		Managed:       &managed,
		Properties:    properties,
	}

	if description != "" {
		trunkBase.Description = &description
	}

	sdkConfig := meta.(*providerMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	log.Printf("Creating trunk base settings %s", name)
	trunkBaseSettings, _, err := edgesAPI.PostTelephonyProvidersEdgesTrunkbasesettings(trunkBase)
	if err != nil {
		return diag.Errorf("Failed to create trunk base settings %s: %s", name, err)
	}

	d.SetId(*trunkBaseSettings.Id)

	log.Printf("Created trunk base settings %s", *trunkBaseSettings.Id)

	return readTrunkBaseSettings(ctx, d, meta)
}

func updateTrunkBaseSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	trunkMetaBase := buildSdkDomainEntityRef(d, "trunk_meta_base_id")
	properties := buildSdkTrunkBaseSettingsProperties(d)
	trunkType := d.Get("trunk_type").(string)
	managed := d.Get("managed").(bool)
	id := d.Id()

	trunkBase := platformclientv2.Trunkbase{
		Id:            &id,
		Name:          &name,
		TrunkMetabase: trunkMetaBase,
		TrunkType:     &trunkType,
		Managed:       &managed,
		Properties:    properties,
	}

	if description != "" {
		trunkBase.Description = &description
	}

	sdkConfig := meta.(*providerMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	// Get the latest version of the setting
	trunkBaseSettings, resp, getErr := edgesAPI.GetTelephonyProvidersEdgesTrunkbasesetting(d.Id(), true)
	if getErr != nil {
		if resp != nil && resp.StatusCode == 404 {
			return nil
		}
		return diag.Errorf("Failed to read trunk base settings %s: %s", d.Id(), getErr)
	}
	trunkBase.Version = trunkBaseSettings.Version

	log.Printf("Updating trunk base settings %s", name)
	trunkBaseSettings, resp, err := edgesAPI.PutTelephonyProvidersEdgesTrunkbasesetting(d.Id(), trunkBase)
	if err != nil {
		return diag.Errorf("Failed to update trunk base settings %s: %s %v", name, err, resp.String())
	}

	log.Printf("Updated trunk base settings %s", *trunkBaseSettings.Id)

	return readTrunkBaseSettings(ctx, d, meta)
}

func readTrunkBaseSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	log.Printf("Reading trunk base settings %s", d.Id())
	trunkBaseSettings, resp, getErr := edgesAPI.GetTelephonyProvidersEdgesTrunkbasesetting(d.Id(), true)

	if getErr != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read trunk base settings %s: %s", d.Id(), getErr)
	}

	d.Set("name", *trunkBaseSettings.Name)
	d.Set("state", *trunkBaseSettings.State)
	if trunkBaseSettings.Description != nil {
		d.Set("description", *trunkBaseSettings.Description)
	}
	if trunkBaseSettings.Managed != nil {
		d.Set("managed", *trunkBaseSettings.Managed)
	}
	if trunkBaseSettings.TrunkMetabase != nil {
		d.Set("trunk_meta_base_id", *trunkBaseSettings.TrunkMetabase.Id)
	}
	d.Set("trunk_type", *trunkBaseSettings.TrunkType)

	d.Set("properties", nil)
	if trunkBaseSettings.Properties != nil {
		d.Set("properties", flattenTrunkBaseSettingsProperties(trunkBaseSettings.Properties))
	}

	log.Printf("Read trunk base settings %s %s", d.Id(), *trunkBaseSettings.Name)

	return nil
}

func flattenTrunkBaseSettingsProperty(property map[string]interface{}) interface{} {
	value := property["value"].(map[string]interface{})
	return value["instance"]
}

func flattenTrunkBaseSettingsProperties(properties interface{}) *schema.Set {
	propertyMap := make(map[string]interface{})

	propertyV := *(properties.(*map[string]interface{}))

	if property, ok := propertyV["trunk_type"].(map[string]interface{}); ok {
		if flattenedProperty := flattenTrunkBaseSettingsProperty(property); flattenedProperty != nil {
			propertyMap[propertiesTrunk_type] = flattenedProperty.(string)
		}
	}
	if property, ok := propertyV["trunk_label"].(map[string]interface{}); ok {
		if flattenedProperty := flattenTrunkBaseSettingsProperty(property); flattenedProperty != nil {
			propertyMap[propertiesTrunk_label] = flattenedProperty.(string)
		}
	}
	if property, ok := propertyV["trunk_enabled"].(map[string]interface{}); ok {
		if flattenedProperty := flattenTrunkBaseSettingsProperty(property); flattenedProperty != nil {
			propertyMap[propertiesTrunk_enabled] = flattenedProperty.(bool)
		}
	}
	if property, ok := propertyV["trunk_maxDialTimeout"].(map[string]interface{}); ok {
		if flattenedProperty := flattenTrunkBaseSettingsProperty(property); flattenedProperty != nil {
			propertyMap[propertiesTrunk_maxDialTimeout] = flattenedProperty.(string)
		}
	}
	if property, ok := propertyV["trunk_maxCallRate"].(map[string]interface{}); ok {
		if flattenedProperty := flattenTrunkBaseSettingsProperty(property); flattenedProperty != nil {
			propertyMap[propertiesTrunk_maxCallRate] = flattenedProperty.(string)
		}
	}
	if property, ok := propertyV["trunk_transport_sipDscpValue"].(map[string]interface{}); ok {
		if flattenedProperty := flattenTrunkBaseSettingsProperty(property); flattenedProperty != nil {
			propertyMap[propertiesTrunk_transport_sipDscpValue] = int(flattenedProperty.(float64))
		}
	}
	if property, ok := propertyV["trunk_transport_tcp_connectTimeout"].(map[string]interface{}); ok {
		if flattenedProperty := flattenTrunkBaseSettingsProperty(property); flattenedProperty != nil {
			propertyMap[propertiesTrunk_transport_tcp_connectTimeout] = int(flattenedProperty.(float64))
		}
	}
	if property, ok := propertyV["trunk_transport_tcp_connectionIdleTimeout"].(map[string]interface{}); ok {
		if flattenedProperty := flattenTrunkBaseSettingsProperty(property); flattenedProperty != nil {
			propertyMap[propertiesTrunk_transport_tcp_connectionIdleTimeout] = int(flattenedProperty.(float64))
		}
	}
	if property, ok := propertyV["trunk_transport_retryableReasonCodes"].(map[string]interface{}); ok {
		if flattenedProperty := flattenTrunkBaseSettingsProperty(property); flattenedProperty != nil {
			propertyMap[propertiesTrunk_transport_retryableReasonCodes] = flattenedProperty.(string)
		}
	}
	if property, ok := propertyV["trunk_transport_retryableCauseCodes"].(map[string]interface{}); ok {
		if flattenedProperty := flattenTrunkBaseSettingsProperty(property); flattenedProperty != nil {
			propertyMap[propertiesTrunk_transport_retryableCauseCodes] = flattenedProperty.(string)
		}
	}
	if property, ok := propertyV["trunk_media_codec"].(map[string]interface{}); ok {
		if flattenedProperty := flattenTrunkBaseSettingsProperty(property); flattenedProperty != nil {
			flattenedPropertySlice := flattenedProperty.([]interface{})
			propertySlice := make([]interface{}, 0)
			for _, flattenedPropertyValue := range flattenedPropertySlice {
				propertySlice = append(propertySlice, flattenedPropertyValue.(interface{}))
			}
			propertyMap[propertiesTrunk_media_codec] = propertySlice
		}
	}
	if property, ok := propertyV["trunk_media_dtmf_method"].(map[string]interface{}); ok {
		if flattenedProperty := flattenTrunkBaseSettingsProperty(property); flattenedProperty != nil {
			propertyMap[propertiesTrunk_media_dtmf_method] = flattenedProperty.(string)
		}
	}
	if property, ok := propertyV["trunk_media_dtmf_payload"].(map[string]interface{}); ok {
		if flattenedProperty := flattenTrunkBaseSettingsProperty(property); flattenedProperty != nil {
			propertyMap[propertiesTrunk_media_dtmf_payload] = int(flattenedProperty.(float64))
		}
	}
	if property, ok := propertyV["trunk_media_dscpValue"].(map[string]interface{}); ok {
		if flattenedProperty := flattenTrunkBaseSettingsProperty(property); flattenedProperty != nil {
			propertyMap[propertiesTrunk_media_dscpValue] = int(flattenedProperty.(float64))
		}
	}
	if property, ok := propertyV["trunk_media_srtpCipherSuites"].(map[string]interface{}); ok {
		if flattenedProperty := flattenTrunkBaseSettingsProperty(property); flattenedProperty != nil {
			flattenedPropertySlice := flattenedProperty.([]interface{})
			propertySlice := make([]interface{}, 0)
			for _, flattenedPropertyValue := range flattenedPropertySlice {
				propertySlice = append(propertySlice, flattenedPropertyValue)
			}
			propertyMap[propertiesTrunk_media_srtpCipherSuites] = propertySlice
		}
	}
	if property, ok := propertyV["trunk_media_disconnectOnIdleRTP"].(map[string]interface{}); ok {
		if flattenedProperty := flattenTrunkBaseSettingsProperty(property); flattenedProperty != nil {
			propertyMap[propertiesTrunk_media_disconnectOnIdleRTP] = flattenedProperty.(bool)
		}
	}
	if property, ok := propertyV["trunk_diagnostic_capture_enabled"].(map[string]interface{}); ok {
		if flattenedProperty := flattenTrunkBaseSettingsProperty(property); flattenedProperty != nil {
			propertyMap[propertiesTrunk_diagnostic_capture_enabled] = flattenedProperty.(bool)
		}
	}
	if property, ok := propertyV["trunk_diagnostic_capture_endTime"].(map[string]interface{}); ok {
		if flattenedProperty := flattenTrunkBaseSettingsProperty(property); flattenedProperty != nil {
			propertyMap[propertiesTrunk_diagnostic_capture_endTime] = flattenedProperty.(string)
		}
	}
	if property, ok := propertyV["trunk_diagnostic_protocol_endTime"].(map[string]interface{}); ok {
		if flattenedProperty := flattenTrunkBaseSettingsProperty(property); flattenedProperty != nil {
			propertyMap[propertiesTrunk_diagnostic_protocol_endTime] = flattenedProperty.(string)
		}
	}
	if property, ok := propertyV["trunk_language"].(map[string]interface{}); ok {
		if flattenedProperty := flattenTrunkBaseSettingsProperty(property); flattenedProperty != nil {
			propertyMap[propertiesTrunk_language] = flattenedProperty.(string)
		}
	}
	propertySet := schema.NewSet(schema.HashResource(trunkBaseSettingsProperties), []interface{}{})
	propertySet.Add(propertyMap)

	return propertySet
}

func deleteTrunkBaseSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	log.Printf("Deleting trunk base settings")
	_, err := edgesAPI.DeleteTelephonyProvidersEdgesTrunkbasesetting(d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete trunk base settings: %s", err)
	}
	log.Printf("Deleted trunk base settings")
	return nil
}

func getAllTrunkBaseSettings(ctx context.Context, sdkConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(ResourceIDMetaMap)

	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	for pageNum := 1; ; pageNum++ {
		trunkBaseSettings, _, getErr := edgesAPI.GetTelephonyProvidersEdgesTrunkbasesettings(pageNum, 100, "", "", false, false, false, []string{"properties"}, "")
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of trunk base settings: %v", getErr)
		}

		if trunkBaseSettings.Entities == nil || len(*trunkBaseSettings.Entities) == 0 {
			break
		}

		for _, trunkBaseSetting := range *trunkBaseSettings.Entities {
			if *trunkBaseSetting.State != "deleted" {
				resources[*trunkBaseSetting.Id] = &ResourceMeta{Name: *trunkBaseSetting.Name}
			}
		}
	}

	return resources, nil
}

func trunkBaseSettingsExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllTrunkBaseSettings),
		RefAttrs:         map[string]*RefAttrSettings{},
	}
}
