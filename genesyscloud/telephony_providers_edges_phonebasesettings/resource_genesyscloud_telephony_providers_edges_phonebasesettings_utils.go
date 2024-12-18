package telephony_providers_edges_phonebasesettings

import (
	"context"
	"fmt"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v149/platformclientv2"
)

func generatePhoneBaseSettingsDataSource(
	resourceLabel string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_telephony_providers_edges_phonebasesettings" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceLabel, name, dependsOnResource)
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
	phoneBaseSettingsResourceLabel,
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
	`, phoneBaseSettingsResourceLabel, name, description, phoneMetaBaseId, strings.Join(otherAttrs, "\n"))
}

func customizePhoneBaseSettingsPropertiesDiff(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	// Defaults must be set on missing properties
	if !diff.NewValueKnown("properties") {
		// properties value not yet in final state. Nothing to do.
		return nil
	}

	id := diff.Id()
	if id == "" {
		return nil
	}

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	phoneBaseProxy := getPhoneBaseProxy(sdkConfig)

	// Retrieve defaults from the settings
	phoneBaseSetting, resp, getErr := phoneBaseProxy.getPhoneBaseSetting(ctx, id)
	if getErr != nil {
		if util.IsStatus404(resp) {
			return nil
		}
		return fmt.Errorf("failed to read phone base settings %s: %s", id, getErr)
	}

	return util.ApplyPropertyDefaults(diff, phoneBaseSetting.Properties)
}

func BuildTelephonyLineBaseProperties(d *schema.ResourceData) *map[string]interface{} {

	if lineBase := d.Get("line_base").([]interface{}); len(lineBase) > 0 {

		lineBaseMap := lineBase[0].(map[string]interface{})

		properties := map[string]interface{}{
			"station_persistent_enabled": &map[string]interface{}{
				"value": &map[string]interface{}{
					"instance": lineBaseMap["station_persistent_enabled"].(bool),
				},
			},
			"station_persistent_timeout": &map[string]interface{}{
				"value": &map[string]interface{}{
					"instance": lineBaseMap["station_persistent_timeout"].(int),
				},
			},
		}
		return &properties
	}
	return nil
}

func flattenTelephonyLineBaseProperties(lineBase *[]platformclientv2.Linebase) []interface{} {
	if lineBase == nil || len(*lineBase) == 0 {
		return nil
	}

	lineBaseMap := make(map[string]interface{})
	propertiesObject := (*lineBase)[0].Properties
	if propertiesObject == nil {
		return []interface{}{lineBaseMap}
	}
	if enabledKey, ok := (*propertiesObject)["station_persistent_enabled"].(map[string]interface{}); ok && enabledKey != nil {
		enabledValue := enabledKey["value"].(map[string]interface{})["instance"]
		if enabledValue != nil {
			resourcedata.SetMapValueIfNotNil(lineBaseMap, "station_persistent_enabled", &enabledValue)
		}
	}
	if timeOutKey, ok := (*propertiesObject)["station_persistent_timeout"].(map[string]interface{}); ok && timeOutKey != nil {
		timeOutKey := timeOutKey["value"].(map[string]interface{})["instance"]
		if timeOutKey != nil {
			resourcedata.SetMapValueIfNotNil(lineBaseMap, "station_persistent_timeout", &timeOutKey)
		}
	}
	return []interface{}{lineBaseMap}
}
