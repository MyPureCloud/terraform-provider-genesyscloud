package integration_action_draft

import (
	"encoding/json"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v154/platformclientv2"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
)

const (
	reqTemplateFileName     = "requesttemplate.vm"
	successTemplateFileName = "successtemplate.vm"
)

func buildActionDraftFromResourceData(d *schema.ResourceData) *platformclientv2.Postactioninput {
	return &platformclientv2.Postactioninput{
		Name:          platformclientv2.String(d.Get("name").(string)),
		Category:      platformclientv2.String(d.Get("category").(string)),
		IntegrationId: platformclientv2.String(d.Get("integration_id").(string)),
		Secure:        platformclientv2.Bool(d.Get("secure").(bool)),
		Contract:      BuildDraftContract(d),
		Config:        buildSdkActionConfig(d),
	}
}

// buildSdkActionConfig takes the resource data and builds the SDK platformclientv2.Actionconfig from it
func buildSdkActionConfig(d *schema.ResourceData) *platformclientv2.Actionconfig {
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

func BuildDraftContract(d *schema.ResourceData) *platformclientv2.Actioncontractinput {
	// Get the contract map first
	contract := d.Get("contract").([]interface{})
	contractMap := contract[0].(map[string]interface{})

	// Extract input and output from the contract map
	contractInput, ok := contractMap["contract_input"].(string)
	if !ok {
		log.Printf("Error: contract_input is not a string")
		return nil
	}
	contractOutput, ok := contractMap["contract_output"].(string)
	if !ok {
		log.Printf("Error: contract_output is not a string")
		return nil
	}

	// Parse input schema with proper error handling
	var inputSchema platformclientv2.Jsonschemadocument
	err := json.Unmarshal([]byte(contractInput), &inputSchema)
	if err != nil {
		log.Printf("Error parsing input schema: %v", err)
		return nil
	}

	// Parse output schema with proper error handling
	var outputSchema platformclientv2.Jsonschemadocument
	err = json.Unmarshal([]byte(contractOutput), &outputSchema)
	if err != nil {
		log.Printf("Error parsing output schema: %v", err)
		return nil
	}

	return &platformclientv2.Actioncontractinput{
		Input: &platformclientv2.Postinputcontract{
			InputSchema: &inputSchema,
		},
		Output: &platformclientv2.Postoutputcontract{
			SuccessSchema: &outputSchema,
		},
	}
}

// flattenActionDraftContract converts the custom ActionContract into a JSON-encoded string
func flattenActionDraftContract(contract platformclientv2.Actioncontract) ([]interface{}, diag.Diagnostics) {
	log.Println("Starting Flatten", contract)
	contractMap := make(map[string]interface{})

	log.Println(contract.String())
	log.Println("Contract Input String", contract.Input.String())
	a := contract.Input.InputSchema.String()
	log.Println("a", a)
	b := contract.Output.SuccessSchema.String()
	contractMap["contract_input"] = a
	contractMap["contract_output"] = b

	return []interface{}{contractMap}, nil
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

// FlattenActionConfigResponse converts the the platformclientv2.Responseconfig into a map
func FlattenActionConfigResponse(sdkResponse platformclientv2.Responseconfig) []interface{} {
	responseMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(responseMap, "translation_map", sdkResponse.TranslationMap)
	resourcedata.SetMapValueIfNotNil(responseMap, "translation_map_defaults", sdkResponse.TranslationMapDefaults)
	resourcedata.SetMapValueIfNotNil(responseMap, "success_template", sdkResponse.SuccessTemplate)

	return []interface{}{responseMap}
}

func buildActionDraftFromResourceDataForUpdate(d *schema.ResourceData, version *int) *platformclientv2.Updatedraftinput {
	log.Println(d.State().String())
	return &platformclientv2.Updatedraftinput{
		Name:     platformclientv2.String(d.Get("name").(string)),
		Category: platformclientv2.String(d.Get("category").(string)),
		Version:  version,
		Secure:   platformclientv2.Bool(d.Get("secure").(bool)),
		Contract: BuildDraftContract(d),
		Config:   buildSdkActionConfig(d),
	}
}
