package util

import (
	"context"
	"encoding/json"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func BuildTelephonyProperties(d *schema.ResourceData) *map[string]interface{} {
	returnValue := make(map[string]interface{})

	if properties := d.Get("properties"); properties != nil {
		inputVal, err := JsonStringToInterface(properties.(string))
		if err != nil {
			return nil
		}
		returnValue = inputVal.(map[string]interface{})
	}

	return &returnValue
}

func FlattenTelephonyProperties(properties interface{}) (string, diag.Diagnostics) {
	if properties == nil {
		return "", nil
	}
	propertiesBytes, err := json.Marshal(properties)
	if err != nil {
		return "", diag.Errorf("Error marshalling properties %v: %v", properties, err)
	}
	return string(propertiesBytes), nil
}

func CustomizePhoneBaseSettingsPropertiesDiff(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
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
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	// Retrieve defaults from the settings
	phoneBaseSetting, resp, getErr := edgesAPI.GetTelephonyProvidersEdgesPhonebasesetting(id)
	if getErr != nil {
		if IsStatus404(resp) {
			return nil
		}
		return fmt.Errorf("failed to read phone base settings %s: %s", id, getErr)
	}

	return applyPropertyDefaults(diff, phoneBaseSetting.Properties)
}

func CustomizeTrunkBaseSettingsPropertiesDiff(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
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
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	// Retrieve defaults from the settings
	trunkBaseSetting, resp, getErr := edgesAPI.GetTelephonyProvidersEdgesTrunkbasesetting(id, true)
	if getErr != nil {
		if IsStatus404(resp) {
			return nil
		}
		return fmt.Errorf("failed to read phone base settings %s: %s", id, getErr)
	}

	return applyPropertyDefaults(diff, trunkBaseSetting.Properties)
}

func applyPropertyDefaults(diff *schema.ResourceDiff, properties *map[string]interface{}) error {
	// Parse resource properties into map
	propertiesJson := diff.Get("properties").(string)
	configMap := map[string]interface{}{}
	if propertiesJson == "" {
		propertiesJson = "{}" // empty object by default
	}
	if err := json.Unmarshal([]byte(propertiesJson), &configMap); err != nil {
		return fmt.Errorf("failure to parse properties for %s: %s", diff.Id(), err)
	}

	// For each property in the schema, check if a value is set in the config
	if properties != nil {
		for name, prop := range *properties {
			if _, set := configMap[name]; !set {
				// Just set a default value if the property wasn't specified
				configMap[name] = prop
			} else {
				configMapProp := configMap[name].(map[string]interface{})
				// Get the instance value from the config
				instance := configMapProp["value"].(map[string]interface{})["instance"]
				if instance == nil {
					continue
				}

				// Assign the property from the API to the config
				configMap[name] = prop
				// Overwrite the instance because that's all we need to set
				configMap[name].(map[string]interface{})["value"].(map[string]interface{})["instance"] = instance
			}
		}
	}

	// Marshal back to string and set as the diff value
	result, err := json.Marshal(configMap)
	if err != nil {
		return fmt.Errorf("failure to marshal properties for %s: %s", diff.Id(), err)
	}

	return diff.SetNew("properties", string(result))
}

func CustomizePhonePropertiesDiff(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
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
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)
	// Retrieve defaults from the settings
	phone, resp, getErr := edgesAPI.GetTelephonyProvidersEdgesPhone(id)
	if getErr != nil {
		if IsStatus404(resp) {
			return nil
		}
		return fmt.Errorf("failed to read phone %s: %s", id, getErr)
	}
	return applyPropertyDefaults(diff, phone.Properties)
}
