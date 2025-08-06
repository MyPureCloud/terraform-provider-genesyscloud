package integration_action_draft

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
)

const (
	reqTemplateFileName     = "requesttemplate.vm"
	successTemplateFileName = "successtemplate.vm"
)

func buildActionDraftFromResourceData(d *schema.ResourceData) *platformclientv2.Postactioninput {
	contract, err := buildDraftContract(d)
	if err != nil {
		log.Fatalf("Error building contract: %v", err)
	}

	return &platformclientv2.Postactioninput{
		Name:          platformclientv2.String(d.Get("name").(string)),
		Category:      platformclientv2.String(d.Get("category").(string)),
		IntegrationId: platformclientv2.String(d.Get("integration_id").(string)),
		Secure:        platformclientv2.Bool(d.Get("secure").(bool)),
		Contract:      contract,
		Config:        buildSdkActionConfig(d),
	}
}

// buildSdkActionConfig takes the resource data and builds the SDK platformclientv2.Actionconfig from it
func buildSdkActionConfig(d *schema.ResourceData) *platformclientv2.Actionconfig {
	ConfigTimeoutSeconds := d.Get("config_timeout_seconds").(int)
	ActionConfig := &platformclientv2.Actionconfig{
		Request:  buildSdkActionConfigRequest(d),
		Response: buildSdkActionConfigResponse(d),
	}

	if ConfigTimeoutSeconds > 0 {
		ActionConfig.TimeoutSeconds = &ConfigTimeoutSeconds
	}
	return ActionConfig
}

// buildSdkActionConfigRequest takes the resource data and builds the SDK platformclientv2.Requestconfig from it
func buildSdkActionConfigRequest(d *schema.ResourceData) *platformclientv2.Requestconfig {
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
func buildSdkActionConfigResponse(d *schema.ResourceData) *platformclientv2.Responseconfig {
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

func buildDraftContract(d *schema.ResourceData) (*platformclientv2.Actioncontractinput, diag.Diagnostics) {
	configInput := d.Get("contract_input").(string)
	configOutput := d.Get("contract_output").(string)

	// Parse input schema with proper error handling
	var inputSchema platformclientv2.Jsonschemadocument
	err := json.Unmarshal([]byte(configInput), &inputSchema)
	if err != nil {
		return nil, util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Failed to parse contract input %s", configInput), err)
	}

	// Parse output schema with proper error handling
	var outputSchema platformclientv2.Jsonschemadocument
	err = json.Unmarshal([]byte(configOutput), &outputSchema)
	if err != nil {
		return nil, util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Failed to parse contract output %s", configOutput), err)
	}

	return &platformclientv2.Actioncontractinput{
		Input: &platformclientv2.Postinputcontract{
			InputSchema: &inputSchema,
		},
		Output: &platformclientv2.Postoutputcontract{
			SuccessSchema: &outputSchema,
		},
	}, nil
}

// flattenActionDraftContract converts the custom ActionContract into a JSON-encoded string
func flattenActionDraftContract(schema platformclientv2.Jsonschemadocument) (string, diag.Diagnostics) {
	schemaBytes, err := json.Marshal(schema)
	if err != nil {
		return "", util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Error marshalling action contract %v", schema), err)
	}
	return string(schemaBytes), nil
}

// FlattenActionConfigRequest converts the platformclientv2.Requestconfig into a map
func flattenActionConfigRequest(sdkRequest platformclientv2.Requestconfig) []interface{} {
	requestMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(requestMap, "request_url_template", sdkRequest.RequestUrlTemplate)
	resourcedata.SetMapValueIfNotNil(requestMap, "request_type", sdkRequest.RequestType)
	resourcedata.SetMapValueIfNotNil(requestMap, "request_template", sdkRequest.RequestTemplate)
	resourcedata.SetMapValueIfNotNil(requestMap, "headers", sdkRequest.Headers)

	return []interface{}{requestMap}
}

// FlattenActionConfigResponse converts the the platformclientv2.Responseconfig into a map
func flattenActionConfigResponse(sdkResponse platformclientv2.Responseconfig) []interface{} {
	responseMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(responseMap, "translation_map", sdkResponse.TranslationMap)
	resourcedata.SetMapValueIfNotNil(responseMap, "translation_map_defaults", sdkResponse.TranslationMapDefaults)
	resourcedata.SetMapValueIfNotNil(responseMap, "success_template", sdkResponse.SuccessTemplate)

	return []interface{}{responseMap}
}
