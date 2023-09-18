package integration_action

import (
	"encoding/json"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

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

func buildSdkActionContract(d *schema.ResourceData) (*ActionContract, diag.Diagnostics) {
	configInput := d.Get("contract_input").(string)
	inputVal, err := gcloud.JsonStringToInterface(configInput)
	if err != nil {
		return nil, diag.Errorf("Failed to parse contract input %s: %v", configInput, err)
	}

	configOutput := d.Get("contract_output").(string)
	outputVal, err := gcloud.JsonStringToInterface(configOutput)
	if err != nil {
		return nil, diag.Errorf("Failed to parse contract output %s: %v", configInput, err)
	}

	return &ActionContract{
		Input:  &ActionInput{InputSchema: &inputVal},
		Output: &ActionOutput{SuccessSchema: &outputVal},
	}, nil
}

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

func flattenActionContract(schema interface{}) (string, diag.Diagnostics) {
	if schema == nil {
		return "", nil
	}
	schemaBytes, err := json.Marshal(schema)
	if err != nil {
		return "", diag.Errorf("Error marshalling action contract %v: %v", schema, err)
	}
	return string(schemaBytes), nil
}

func flattenActionConfigRequest(sdkRequest platformclientv2.Requestconfig) []interface{} {
	requestMap := make(map[string]interface{})
	if sdkRequest.RequestUrlTemplate != nil {
		requestMap["request_url_template"] = *sdkRequest.RequestUrlTemplate
	}
	if sdkRequest.RequestType != nil {
		requestMap["request_type"] = *sdkRequest.RequestType
	}
	if sdkRequest.RequestTemplate != nil {
		requestMap["request_template"] = *sdkRequest.RequestTemplate
	}
	if sdkRequest.Headers != nil {
		requestMap["headers"] = *sdkRequest.Headers
	}
	return []interface{}{requestMap}
}

func flattenActionConfigResponse(sdkResponse platformclientv2.Responseconfig) []interface{} {
	responseMap := make(map[string]interface{})
	if sdkResponse.TranslationMap != nil {
		responseMap["translation_map"] = *sdkResponse.TranslationMap
	}
	if sdkResponse.TranslationMapDefaults != nil {
		responseMap["translation_map_defaults"] = *sdkResponse.TranslationMapDefaults
	}
	if sdkResponse.SuccessTemplate != nil {
		responseMap["success_template"] = *sdkResponse.SuccessTemplate
	}
	return []interface{}{responseMap}
}
