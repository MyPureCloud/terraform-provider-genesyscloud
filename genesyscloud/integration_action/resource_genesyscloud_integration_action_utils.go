package integration_action

import (
	"encoding/json"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"

	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
)

/*
The resource_genesyscloud_integration_action_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.

Note:  Look for opportunities to minimize boilerplate code using functions and Generics
*/

const (
	reqTemplateFileName     = "requesttemplate.vm"
	successTemplateFileName = "successtemplate.vm"
)

type ActionInput struct {
	InputSchema *interface{} `json:"inputSchema,omitempty"`
}
type ActionOutput struct {
	SuccessSchema *interface{} `json:"successSchema,omitempty"`
}

type ActionContract struct {
	Output *ActionOutput `json:"output,omitempty"`
	Input  *ActionInput  `json:"input,omitempty"`
}

type IntegrationAction struct {
	Id            *string                        `json:"id,omitempty"`
	Name          *string                        `json:"name,omitempty"`
	Category      *string                        `json:"category,omitempty"`
	IntegrationId *string                        `json:"integrationId,omitempty"`
	Secure        *bool                          `json:"secure,omitempty"`
	Config        *platformclientv2.Actionconfig `json:"config,omitempty"`
	Contract      *ActionContract                `json:"contract,omitempty"`
	Version       *int                           `json:"version,omitempty"`
}

// BuildSdkActionContract takes the resource data and builds the custom ActionContract from it
func BuildSdkActionContract(d *schema.ResourceData) (*ActionContract, diag.Diagnostics) {
	configInput := d.Get("contract_input").(string)
	inputVal, err := util.JsonStringToInterface(configInput)
	if err != nil {
		return nil, util.BuildDiagnosticError(resourceName, fmt.Sprintf("Failed to parse contract input %s", configInput), err)
	}

	configOutput := d.Get("contract_output").(string)
	outputVal, err := util.JsonStringToInterface(configOutput)
	if err != nil {
		return nil, util.BuildDiagnosticError(resourceName, fmt.Sprintf("Failed to parse contract output %s", configInput), err)
	}

	return &ActionContract{
		Input:  &ActionInput{InputSchema: &inputVal},
		Output: &ActionOutput{SuccessSchema: &outputVal},
	}, nil
}

// buildSdkActionConfig takes the resource data and builds the SDK platformclientv2.Actionconfig from it
func BuildSdkActionConfig(d *schema.ResourceData) *platformclientv2.Actionconfig {
	ConfigTimeoutSeconds := d.Get("config_timeout_seconds").(int)
	ActionConfig := &platformclientv2.Actionconfig{
		Request:  BuildSdkActionConfigRequest(d),
		Response: BuildSdkActionConfigResponse(d),
	}

	if ConfigTimeoutSeconds > 0 {
		ActionConfig.TimeoutSeconds = &ConfigTimeoutSeconds
	}

	return ActionConfig
}

// buildSdkActionConfigRequest takes the resource data and builds the SDK platformclientv2.Requestconfig from it
func BuildSdkActionConfigRequest(d *schema.ResourceData) *platformclientv2.Requestconfig {
	if configRequest := d.Get("config_request"); configRequest != nil {
		if configList := configRequest.([]interface{}); len(configList) > 0 {
			configMap := configList[0].(map[string]interface{})

			urlTemplate := configMap["request_url_template"].(string)
			template := configMap["request_template"].(string)
			reqType := configMap["request_type"].(string)
			headers := map[string]string{}
			if headerVal, ok := configMap["headers"]; ok && headerVal != nil {
				for key, val := range headerVal.(map[string]interface{}) {
					headers[key] = val.(string)
				}
			}

			return &platformclientv2.Requestconfig{
				RequestUrlTemplate: &urlTemplate,
				RequestTemplate:    &template,
				RequestType:        &reqType,
				Headers:            &headers,
			}
		}
	}
	return &platformclientv2.Requestconfig{}
}

// buildSdkActionConfigResponse takes the resource data and builds the SDK platformclientv2.Responseconfig from it
func BuildSdkActionConfigResponse(d *schema.ResourceData) *platformclientv2.Responseconfig {
	if configResponse := d.Get("config_response"); configResponse != nil {
		if configList := configResponse.([]interface{}); len(configList) > 0 {
			configMap := configList[0].(map[string]interface{})

			transMap := map[string]string{}
			if mapVal, ok := configMap["translation_map"]; ok && mapVal != nil {
				for key, val := range mapVal.(map[string]interface{}) {
					transMap[key] = val.(string)
				}
			}
			transMapDefaults := map[string]string{}
			if mapVal, ok := configMap["translation_map_defaults"]; ok && mapVal != nil {
				for key, val := range mapVal.(map[string]interface{}) {
					transMapDefaults[key] = val.(string)
				}
			}
			var successTemplate string
			if tempVal, ok := configMap["success_template"]; ok {
				successTemplate = tempVal.(string)
			}

			return &platformclientv2.Responseconfig{
				TranslationMap:         &transMap,
				TranslationMapDefaults: &transMapDefaults,
				SuccessTemplate:        &successTemplate,
			}
		}
	}
	return &platformclientv2.Responseconfig{}
}

// flattenActionContract converts the custom ActionContract into a JSON-encoded string
func flattenActionContract(schema interface{}) (string, diag.Diagnostics) {
	if schema == nil {
		return "", nil
	}
	schemaBytes, err := json.Marshal(schema)
	if err != nil {
		return "", util.BuildDiagnosticError(resourceName, fmt.Sprintf("Error marshalling action contract %v", schema), err)
	}
	return string(schemaBytes), nil
}

// flattenActionConfigRequest converts the platformclientv2.Requestconfig into a map
func FlattenActionConfigRequest(sdkRequest platformclientv2.Requestconfig) []interface{} {
	requestMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(requestMap, "request_url_template", sdkRequest.RequestUrlTemplate)
	resourcedata.SetMapValueIfNotNil(requestMap, "request_type", sdkRequest.RequestType)
	resourcedata.SetMapValueIfNotNil(requestMap, "request_template", sdkRequest.RequestTemplate)
	resourcedata.SetMapValueIfNotNil(requestMap, "headers", sdkRequest.Headers)

	return []interface{}{requestMap}
}

// FlattenActionConfigResponse converts the the platformclientv2.Responseconfig into a map
func FlattenActionConfigResponse(sdkResponse platformclientv2.Responseconfig) []interface{} {
	responseMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(responseMap, "translation_map", sdkResponse.TranslationMap)
	resourcedata.SetMapValueIfNotNil(responseMap, "translation_map_defaults", sdkResponse.TranslationMapDefaults)
	resourcedata.SetMapValueIfNotNil(responseMap, "success_template", sdkResponse.SuccessTemplate)

	return []interface{}{responseMap}
}
