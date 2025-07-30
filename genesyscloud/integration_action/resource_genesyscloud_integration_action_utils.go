package integration_action

import (
	"encoding/json"
	"fmt"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
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
		return nil, util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Failed to parse contract input %s", configInput), err)
	}

	configOutput := d.Get("contract_output").(string)
	outputVal, err := util.JsonStringToInterface(configOutput)
	if err != nil {
		return nil, util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Failed to parse contract output %s", configInput), err)
	}

	return &ActionContract{
		Input:  &ActionInput{InputSchema: &inputVal},
		Output: &ActionOutput{SuccessSchema: &outputVal},
	}, nil
}

// BuildSdkActionContract takes the resource data and builds the custom ActionContract from it
func BuildSdkActionContractInput(d *schema.ResourceData) (*platformclientv2.Actioncontractinput, diag.Diagnostics) {
	configInput := d.Get("contract_input").(string)

	// Parse input schema with proper error handling
	var inputSchema platformclientv2.Jsonschemadocument
	err := json.Unmarshal([]byte(configInput), &inputSchema)
	if err != nil {
		return nil, util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Failed to parse contract input %s", configInput), err)
	}

	configOutput := d.Get("contract_output").(string)

	// Parse output schema with proper error handling
	var outputSchema platformclientv2.Jsonschemadocument
	err = json.Unmarshal([]byte(configOutput), &outputSchema)
	if err != nil {
		return nil, util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Failed to parse contract output %s", configOutput), err)
	}

	return &platformclientv2.Actioncontractinput{
		Input:  &platformclientv2.Postinputcontract{InputSchema: &inputSchema},
		Output: &platformclientv2.Postoutputcontract{SuccessSchema: &outputSchema},
	}, nil
}

// BuildSdkActionConfig takes the resource data and builds the SDK platformclientv2.Actionconfig from it
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

// BuildSdkActionConfigRequest takes the resource data and builds the SDK platformclientv2.Requestconfig from it
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

// BuildSdkActionConfigResponse takes the resource data and builds the SDK platformclientv2.Responseconfig from it
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
		return "", util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Error marshalling action contract %v", schema), err)
	}
	return string(schemaBytes), nil
}

// FlattenActionConfigRequest converts the platformclientv2.Requestconfig into a map
func FlattenActionConfigRequest(sdkRequest platformclientv2.Requestconfig) []interface{} {
	requestMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(requestMap, "request_url_template", sdkRequest.RequestUrlTemplate)
	resourcedata.SetMapValueIfNotNil(requestMap, "request_type", sdkRequest.RequestType)
	resourcedata.SetMapValueIfNotNil(requestMap, "request_template", sdkRequest.RequestTemplate)
	resourcedata.SetMapValueIfNotNil(requestMap, "headers", sdkRequest.Headers)

	return []interface{}{requestMap}
}

// FlattenActionConfigResponse converts the platformclientv2.Responseconfig into a map
func FlattenActionConfigResponse(sdkResponse platformclientv2.Responseconfig) []interface{} {
	responseMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(responseMap, "translation_map", sdkResponse.TranslationMap)
	resourcedata.SetMapValueIfNotNil(responseMap, "translation_map_defaults", sdkResponse.TranslationMapDefaults)
	resourcedata.SetMapValueIfNotNil(responseMap, "success_template", sdkResponse.SuccessTemplate)

	return []interface{}{responseMap}
}

// FlattenFunctionConfigRequest converts the platformclientv2.Functionconfig into a map
func FlattenFunctionConfigRequest(functionConfig platformclientv2.Functionconfig) []interface{} {
	functionMap := make(map[string]interface{})

	// Extract function settings from the Function field
	if functionConfig.Function != nil {
		resourcedata.SetMapValueIfNotNil(functionMap, "description", functionConfig.Function.Description)
		resourcedata.SetMapValueIfNotNil(functionMap, "handler", functionConfig.Function.Handler)
		resourcedata.SetMapValueIfNotNil(functionMap, "runtime", functionConfig.Function.Runtime)
		resourcedata.SetMapValueIfNotNil(functionMap, "timeout_seconds", functionConfig.Function.TimeoutSeconds)
		resourcedata.SetMapValueIfNotNil(functionMap, "zip_id", functionConfig.Function.ZipId)
	}

	return []interface{}{functionMap}
}

// BuildSdkFunctionConfig takes the resource data and builds the SDK platformclientv2.Functionconfig from it
func BuildSdkFunctionConfig(d *schema.ResourceData) *platformclientv2.Functionconfig {
	if functionConfig := d.Get("function_config"); functionConfig != nil {
		if configList := functionConfig.([]interface{}); len(configList) > 0 {
			configMap := configList[0].(map[string]interface{})

			// Extract function settings
			var description string
			if descVal, ok := configMap["description"]; ok && descVal != nil {
				description = descVal.(string)
			}

			var handler string
			if handlerVal, ok := configMap["handler"]; ok && handlerVal != nil {
				handler = handlerVal.(string)
			}

			var runtime string
			if runtimeVal, ok := configMap["runtime"]; ok && runtimeVal != nil {
				runtime = runtimeVal.(string)
			}

			var timeoutSeconds int
			if timeoutVal, ok := configMap["timeout_seconds"]; ok && timeoutVal != nil {
				timeoutSeconds = timeoutVal.(int)
			}

			var zipId string
			if zipIdVal, ok := configMap["zip_id"]; ok && zipIdVal != nil {
				zipId = zipIdVal.(string)
			}

			// Create the Function object
			function := &platformclientv2.Function{
				Description:    platformclientv2.String(description),
				Handler:        platformclientv2.String(handler),
				Runtime:        platformclientv2.String(runtime),
				TimeoutSeconds: platformclientv2.Int(timeoutSeconds),
				ZipId:          platformclientv2.String(zipId),
			}

			// Create the Functionconfig object
			return &platformclientv2.Functionconfig{
				Function: function,
			}
		}
	}
	return &platformclientv2.Functionconfig{}
}
