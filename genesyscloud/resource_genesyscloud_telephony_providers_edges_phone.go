package genesyscloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v46/platformclientv2"
	"log"
)

func resourcePhone() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Phone",

		CreateContext: createWithPooledClient(createPhone),
		ReadContext:   readWithPooledClient(readPhone),
		UpdateContext: updateWithPooledClient(updatePhone),
		DeleteContext: deleteWithPooledClient(deletePhone),
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
				Description:  "Indicates if the resource is active, inactive, or deleted. Valid values: active, inactive, deleted.",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "active",
				ValidateFunc: validation.StringInSlice([]string{"active", "inactive", "deleted"}, false),
			},
			"site_id": {
				Description: "The site ID associated to the phone.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"phone_base_settings_id": {
				Description: "Phone Base Settings ID.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"line_base_settings_id": {
				Description: "Line Base Settings ID.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"phone_meta_base_id": {
				Description: "Phone Meta Base ID.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"web_rtc_user_id": {
				Description: "Web RTC User ID.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"capabilities": {
				Description: "Phone Capabilities.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"provisions": {
							Description: "Provisions",
							Type:        schema.TypeBool,
							Optional:    true,
						},
						"registers": {
							Description: "Registers",
							Type:        schema.TypeBool,
							Optional:    true,
						},
						"dual_registers": {
							Description: "Dual Registers",
							Type:        schema.TypeBool,
							Optional:    true,
						},
						"hardware_id_type": {
							Description: "HardwareId Type",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"allow_reboot": {
							Description: "Allow Reboot",
							Type:        schema.TypeBool,
							Optional:    true,
						},
						"no_rebalance": {
							Description: "No Rebalance",
							Type:        schema.TypeBool,
							Optional:    true,
						},
						"no_cloud_provisioning": {
							Description: "No Cloud Provisioning",
							Type:        schema.TypeBool,
							Optional:    true,
						},
						"media_codecs": {
							Description: "Media Codecs",
							Type:        schema.TypeList,
							Optional:    true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice([]string{"audio/opus", "audio/pcmu", "audio/pcma", "audio/g729", "audio/g722"}, false),
							},
						},
						"cdm": {
							Description: "CDM",
							Type:        schema.TypeBool,
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

func createPhone(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	state := d.Get("state").(string)
	site := buildSdkDomainEntityRef(d, "site_id")
	phoneBaseSettings := buildSdkDomainEntityRef(d, "phone_base_settings_id")
	lineBaseSettings := buildSdkDomainEntityRef(d, "line_base_settings_id")
	phoneMetaBaseId, err := getPhoneMetaBaseId(meta, *phoneBaseSettings.Id)
	if err != nil {
		return diag.Errorf("Failed to get phone meta base for %s: %s", name, err)
	}

	phoneMetaBase := &platformclientv2.Domainentityref{
		Id: &phoneMetaBaseId,
	}
	lineName := "line_" + *lineBaseSettings.Id
	line := &platformclientv2.Line{
		Name:             &lineName,
		LineBaseSettings: lineBaseSettings,
	}
	capabilities := buildSdkCapabilities(d)
	webRtcUserId := d.Get("web_rtc_user_id")

	sdkConfig := meta.(*providerMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	createPhone := &platformclientv2.Phone{
		Name:              &name,
		State:             &state,
		Site:              site,
		PhoneBaseSettings: phoneBaseSettings,
		LineBaseSettings:  lineBaseSettings,
		PhoneMetaBase:     phoneMetaBase,
		Lines: &[]platformclientv2.Line{
			*line,
		},
		Capabilities: capabilities,
	}

	if webRtcUserId != "" {
		createPhone.WebRtcUser = buildSdkDomainEntityRef(d, "web_rtc_user_id")
	}

	log.Printf("Creating phone %s", name)
	phone, _, err := edgesAPI.PostTelephonyProvidersEdgesPhones(*createPhone)
	if err != nil {
		return diag.Errorf("Failed to create phone %s: %s", name, err)
	}

	d.SetId(*phone.Id)

	log.Printf("Created phone %s", *phone.Id)

	return readPhone(ctx, d, meta)
}

func readPhone(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	log.Printf("Reading phone %s", d.Id())
	currentPhone, resp, getErr := edgesAPI.GetTelephonyProvidersEdgesPhone(d.Id())

	if getErr != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read phone %s: %s", d.Id(), getErr)
	}

	d.Set("name", *currentPhone.Name)
	d.Set("state", *currentPhone.State)
	d.Set("site_id", *currentPhone.Site.Id)
	d.Set("phone_base_settings_id", *currentPhone.PhoneBaseSettings.Id)
	d.Set("line_base_settings_id", *currentPhone.LineBaseSettings.Id)
	if currentPhone.PhoneMetaBase != nil {
		d.Set("phone_meta_base_id", *currentPhone.PhoneMetaBase.Id)
	}

	if currentPhone.WebRtcUser != nil {
		d.Set("web_rtc_user_id", *currentPhone.WebRtcUser.Id)
	}

	if currentPhone.Capabilities != nil {
		d.Set("capabilities", flattenPhoneCapabilities(currentPhone.Capabilities))
	}

	log.Printf("Read phone %s %s", d.Id(), *currentPhone.Name)
	return nil
}

func updatePhone(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	site := buildSdkDomainEntityRef(d, "site_id")
	phoneBaseSettings := buildSdkDomainEntityRef(d, "phone_base_settings_id")
	lineBaseSettings := buildSdkDomainEntityRef(d, "line_base_settings_id")
	phoneMetaBase := buildSdkDomainEntityRef(d, "phone_meta_base_id")
	lineName := "line_" + *lineBaseSettings.Id
	line := &platformclientv2.Line{
		Name:             &lineName,
		LineBaseSettings: lineBaseSettings,
	}
	webRtcUserId := d.Get("web_rtc_user_id")

	sdkConfig := meta.(*providerMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	updatePhoneBody := &platformclientv2.Phone{
		Name:              &name,
		Site:              site,
		PhoneBaseSettings: phoneBaseSettings,
		PhoneMetaBase:     phoneMetaBase,
		Lines: &[]platformclientv2.Line{
			*line,
		},
	}

	if webRtcUserId != "" {
		updatePhoneBody.WebRtcUser = buildSdkDomainEntityRef(d, "web_rtc_user_id")
	}

	log.Printf("Updating phone %s", name)
	phone, _, err := edgesAPI.PutTelephonyProvidersEdgesPhone(d.Id(), *updatePhoneBody)
	if err != nil {
		return diag.Errorf("Failed to update phone %s: %s", name, err)
	}

	log.Printf("Updated phone %s", *phone.Id)

	return readPhone(ctx, d, meta)
}

func deletePhone(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	log.Printf("Deleting Phone")
	_, err := edgesAPI.DeleteTelephonyProvidersEdgesPhone(d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete phone: %s", err)
	}
	log.Printf("Deleted Phone")
	return nil
}

func getPhoneMetaBaseId(meta interface{}, phoneBaseSettingsId string) (string, error) {
	sdkConfig := meta.(*providerMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	phoneBase, _, err := edgesAPI.GetTelephonyProvidersEdgesPhonebasesetting(phoneBaseSettingsId)
	if err != nil {
		return "", err
	}

	return *phoneBase.PhoneMetaBase.Id, nil
}

func flattenPhoneCapabilities(capabilities *platformclientv2.Phonecapabilities) []interface{} {
	if capabilities == nil {
		return nil
	}

	capabilitiesMap := make(map[string]interface{})
	if capabilities.Provisions != nil {
		capabilitiesMap["provisions"] = *capabilities.Provisions
	}
	if capabilities.Registers != nil {
		capabilitiesMap["registers"] = *capabilities.Registers
	}
	if capabilities.DualRegisters != nil {
		capabilitiesMap["dual_registers"] = *capabilities.DualRegisters
	}
	if capabilities.HardwareIdType != nil {
		capabilitiesMap["hardware_id_type"] = *capabilities.HardwareIdType
	}
	if capabilities.AllowReboot != nil {
		capabilitiesMap["allow_reboot"] = *capabilities.AllowReboot
	}
	if capabilities.NoRebalance != nil {
		capabilitiesMap["no_rebalance"] = *capabilities.NoRebalance
	}
	if capabilities.NoCloudProvisioning != nil {
		capabilitiesMap["no_cloud_provisioning"] = *capabilities.NoCloudProvisioning
	}
	if capabilities.MediaCodecs != nil {
		capabilitiesMap["media_codecs"] = *capabilities.MediaCodecs
	}
	if capabilities.Cdm != nil {
		capabilitiesMap["cdm"] = *capabilities.Cdm
	}

	return []interface{}{capabilitiesMap}
}

func getAllPhones(ctx context.Context, sdkConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(ResourceIDMetaMap)

	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	for pageNum := 1; ; pageNum++ {
		phones, _, getErr := edgesAPI.GetTelephonyProvidersEdgesPhones(100, pageNum, "", "", "", "", "", "", "", "", "", "", "", "", "", nil, nil)
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of phones: %v", getErr)
		}

		if phones.Entities == nil || len(*phones.Entities) == 0 {
			break
		}

		for _, phone := range *phones.Entities {
			if *phone.State != "deleted" {
				resources[*phone.Id] = &ResourceMeta{Name: *phone.Name}
			}
		}
	}

	return resources, nil
}

func phoneExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllPhones),
		RefAttrs:         map[string]*RefAttrSettings{
			"web_rtc_user_id": {RefType: "genesyscloud_user"},
		},
	}
}

func buildSdkCapabilities(d *schema.ResourceData) *platformclientv2.Phonecapabilities {
	if capabilities := d.Get("capabilities").([]interface{}); capabilities != nil {
		sdkPhoneCapabilities := platformclientv2.Phonecapabilities{}
		if len(capabilities) > 0 {
			capabilitiesMap := capabilities[0].(map[string]interface{})

			// Only set non-empty values.
			provisions := capabilitiesMap["provisions"].(bool)
			registers := capabilitiesMap["registers"].(bool)
			dualRegisters := capabilitiesMap["dual_registers"].(bool)
			var hardwareIdType string
			if checkHardwareIdType := capabilitiesMap["hardware_id_type"].(string); len(checkHardwareIdType) > 0 {
				hardwareIdType = checkHardwareIdType
			}
			allowReboot := capabilitiesMap["allow_reboot"].(bool)
			noRebalance := capabilitiesMap["no_rebalance"].(bool)
			noCloudProvisioning := capabilitiesMap["no_cloud_provisioning"].(bool)
			mediaCodecs := make([]string, 0)
			if checkMediaCodecs := capabilitiesMap["media_codecs"].([]interface{}); len(checkMediaCodecs) > 0 {
				for _, codec := range checkMediaCodecs {
					mediaCodecs = append(mediaCodecs, fmt.Sprintf("%v", codec))
				}
			}
			cdm := capabilitiesMap["cdm"].(bool)

			sdkPhoneCapabilities = platformclientv2.Phonecapabilities{
				Provisions:          &provisions,
				Registers:           &registers,
				DualRegisters:       &dualRegisters,
				HardwareIdType:      &hardwareIdType,
				AllowReboot:         &allowReboot,
				NoRebalance:         &noRebalance,
				NoCloudProvisioning: &noCloudProvisioning,
				MediaCodecs:         &mediaCodecs,
				Cdm:                 &cdm,
			}
		}
		return &sdkPhoneCapabilities
	}
	return nil
}
