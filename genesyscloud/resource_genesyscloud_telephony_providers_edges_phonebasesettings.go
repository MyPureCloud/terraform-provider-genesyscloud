package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

var (
	phoneCapabilities = &schema.Resource{
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
	}
)

func ResourcePhoneBaseSettings() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Phone Base Settings",

		CreateContext: CreateWithPooledClient(createPhoneBaseSettings),
		ReadContext:   ReadWithPooledClient(readPhoneBaseSettings),
		UpdateContext: UpdateWithPooledClient(updatePhoneBaseSettings),
		DeleteContext: DeleteWithPooledClient(deletePhoneBaseSettings),
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
				DiffSuppressFunc: SuppressEquivalentJsonDiffs,
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
	phoneMetaBase := BuildSdkDomainEntityRef(d, "phone_meta_base_id")
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
	phoneMetaBase := BuildSdkDomainEntityRef(d, "phone_meta_base_id")
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
		if IsStatus404(resp) {
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
	return WithRetriesForRead(ctx, d, func() *retry.RetryError {
		phoneBaseSettings, resp, getErr := edgesAPI.GetTelephonyProvidersEdgesPhonebasesetting(d.Id())
		if getErr != nil {
			if IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read phone base settings %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read phone base settings %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourcePhoneBaseSettings())
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
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	log.Printf("Deleting phone base settings")
	_, err := edgesAPI.DeleteTelephonyProvidersEdgesPhonebasesetting(d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete phone base settings: %s", err)
	}

	return WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		phoneBaseSettings, resp, err := edgesAPI.GetTelephonyProvidersEdgesPhonebasesetting(d.Id())
		if err != nil {
			if IsStatus404(resp) {
				// Phone base settings deleted
				log.Printf("Deleted Phone base settings %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting Phone base settings %s: %s", d.Id(), err))
		}

		if phoneBaseSettings.State != nil && *phoneBaseSettings.State == "deleted" {
			// Phone base settings deleted
			log.Printf("Deleted Phone base settings %s", d.Id())
			return nil
		}

		return retry.RetryableError(fmt.Errorf("Phone base settings %s still exists", d.Id()))
	})
}

func getAllPhoneBaseSettings(ctx context.Context, sdkConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)

	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	err := WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			phoneBaseSettings, resp, getErr := edgesAPI.GetTelephonyProvidersEdgesPhonebasesettings(pageSize, pageNum, "", "", nil, "")
			if getErr != nil {
				if IsStatus404(resp) {
					return retry.RetryableError(fmt.Errorf("Failed to get page of phonebasesettings: %v", getErr))
				}
				return retry.NonRetryableError(fmt.Errorf("Failed to get page of phonebasesettings: %v", getErr))
			}

			if phoneBaseSettings.Entities == nil || len(*phoneBaseSettings.Entities) == 0 {
				break
			}

			for _, phoneBaseSetting := range *phoneBaseSettings.Entities {
				if phoneBaseSetting.State != nil && *phoneBaseSetting.State != "deleted" {
					resources[*phoneBaseSetting.Id] = &resourceExporter.ResourceMeta{Name: *phoneBaseSetting.Name}
				}
			}
		}

		return nil
	})

	return resources, err
}

func PhoneBaseSettingsExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc:     GetAllWithPooledClient(getAllPhoneBaseSettings),
		RefAttrs:             map[string]*resourceExporter.RefAttrSettings{},
		JsonEncodeAttributes: []string{"properties"},
	}
}

func buildSdkCapabilities(d *schema.ResourceData) *platformclientv2.Phonecapabilities {
	if capabilities := d.Get("capabilities").([]interface{}); capabilities != nil {
		sdkPhoneCapabilities := platformclientv2.Phonecapabilities{}
		if len(capabilities) > 0 {
			if _, ok := capabilities[0].(map[string]interface{}); !ok {
				return nil
			}
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

func flattenPhoneCapabilities(capabilities *platformclientv2.Phonecapabilities) []interface{} {
	if capabilities == nil {
		return nil
	}

	capabilitiesMap := make(map[string]interface{})
	resourcedata.SetMapValueIfNotNil(capabilitiesMap, "provisions", capabilities.Provisions)
	resourcedata.SetMapValueIfNotNil(capabilitiesMap, "registers", capabilities.Registers)
	resourcedata.SetMapValueIfNotNil(capabilitiesMap, "dual_registers", capabilities.DualRegisters)
	resourcedata.SetMapValueIfNotNil(capabilitiesMap, "hardware_id_type", capabilities.HardwareIdType)
	resourcedata.SetMapValueIfNotNil(capabilitiesMap, "allow_reboot", capabilities.AllowReboot)
	resourcedata.SetMapValueIfNotNil(capabilitiesMap, "no_rebalance", capabilities.NoRebalance)
	resourcedata.SetMapValueIfNotNil(capabilitiesMap, "no_cloud_provisioning", capabilities.NoCloudProvisioning)
	resourcedata.SetMapValueIfNotNil(capabilitiesMap, "media_codecs", capabilities.MediaCodecs)
	resourcedata.SetMapValueIfNotNil(capabilitiesMap, "cdm", capabilities.Cdm)

	return []interface{}{capabilitiesMap}
}

func GeneratePhoneBaseSettingsResourceWithCustomAttrs(
	phoneBaseSettingsRes,
	name,
	description,
	phoneMetaBaseId string,
	otherAttrs ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_telephony_providers_edges_phonebasesettings" "%s" {
		name = "%s"
		description = "%s"
		phone_meta_base_id = "%s"
		%s
	}
	`, phoneBaseSettingsRes, name, description, phoneMetaBaseId, strings.Join(otherAttrs, "\n"))
}
