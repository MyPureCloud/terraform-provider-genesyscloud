package integration_action_draft

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v152/platformclientv2"
	"log"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"
)

func getAllIntegrationActionDrafts(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	iap := getIntegrationActionsProxy(clientConfig)

	actions, resp, err := iap.getAllIntegrationActionDrafts(ctx, "")
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get integration action drafts %s", err), resp)
	}

	for _, action := range *actions {
		// Don't include "static" actions
		if strings.HasPrefix(*action.Id, "static") {
			continue
		}
		resources[*action.Id] = &resourceExporter.ResourceMeta{BlockLabel: *action.Name}
	}
	return resources, nil
}

func createIntegrationActionDraft(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	iap := getIntegrationActionsProxy(sdkConfig)

	log.Printf("Creating integration action draft %s", name)

	draftRequest := buildActionDraftFromResourceData(d)

	draft, resp, err := iap.createIntegrationActionDraft(ctx, *draftRequest)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create integration action draft %s error: %s", name, err), resp)
	}

	d.SetId(*draft.Id)
	log.Printf("Created integration action draft %s %s", name, *draft.Id)

	return readIntegrationActionDraft(ctx, d, meta)
}

func readIntegrationActionDraft(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	iap := getIntegrationActionsProxy(sdkConfig)

	log.Printf("Reading integration action draft %s", d.Id())
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceIntegrationActionDraft(), constants.ConsistencyChecks(), ResourceType)

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		draft, resp, getErr := iap.getIntegrationActionDraftById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read action draft %s: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read action draft %s: %s", d.Id(), getErr), resp))
		}

		// Retrieve config request/response templates
		reqTemp, resp, err := iap.getIntegrationActionDraftTemplate(ctx, d.Id(), reqTemplateFileName)
		if err != nil {
			if util.IsStatus404(resp) {
				d.SetId("")
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read request template for integration action draft %s | error: %s", d.Id(), err), resp))
		}

		successTemp, resp, err := iap.getIntegrationActionDraftTemplate(ctx, d.Id(), successTemplateFileName)
		if err != nil {
			if util.IsStatus404(resp) {
				d.SetId("")
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read success template for integration action draft %s | error: %s", d.Id(), err), resp))
		}

		resourcedata.SetNillableValue(d, "name", draft.Name)
		resourcedata.SetNillableValue(d, "category", draft.Category)
		resourcedata.SetNillableValue(d, "integration_id", draft.IntegrationId)
		resourcedata.SetNillableValue(d, "secure", draft.Secure)
		resourcedata.SetNillableValue(d, "config_timeout_seconds", draft.Config.TimeoutSeconds)

		log.Println(draft.Contract.Input.String())
		if draft.Contract != nil && draft.Contract.Input != nil {
			input, err := flattenActionDraftContract(*draft.Contract.Input)
			if err != nil {
				return retry.NonRetryableError(fmt.Errorf("%v", err))
			}
			_ = d.Set("contract_input", input)
		} else {
			_ = d.Set("contract_input", nil)
		}

		if draft.Contract != nil && draft.Contract.Output != nil && draft.Contract.Output.SuccessSchema != nil {
			output, err := flattenActionDraftContract(*draft.Contract.Output.SuccessSchema)
			if err != nil {
				return retry.NonRetryableError(fmt.Errorf("%v", err))
			}
			_ = d.Set("contract_output", output)
		} else {
			_ = d.Set("contract_output", nil)
		}

		if draft.Config != nil && draft.Config.Request != nil {
			draft.Config.Request.RequestTemplate = reqTemp
			_ = d.Set("config_request", FlattenActionDraftConfigRequest(*draft.Config.Request))
		} else {
			_ = d.Set("config_request", nil)
		}

		if draft.Config != nil && draft.Config.Response != nil {
			draft.Config.Response.SuccessTemplate = successTemp
			_ = d.Set("config_response", FlattenActionDraftConfigResponse(*draft.Config.Response))
		} else {
			_ = d.Set("config_response", nil)
		}

		log.Printf("Read integration action draft %s %s", d.Id(), *draft.Name)
		return cc.CheckState(d)
	})
}

func updateIntegrationActionDraft(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	iap := getIntegrationActionsProxy(sdkConfig)
	name := d.Get("name").(string)

	log.Printf("Updating integration action draft %s", name)

	draftRequest := buildActionDraftFromResourceDataForUpdate(d)

	_, resp, err := iap.updateIntegrationActionDraft(ctx, d.Id(), *draftRequest)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update integration action draft %s error: %s", name, err), resp)
	}

	log.Printf("Updated integration action draft %s", name)
	return readIntegrationActionDraft(ctx, d, meta)
}

// deleteIntegrationActionDraft is used by the integration action resource to delete an action from Genesys cloud.
func deleteIntegrationActionDraft(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	iap := getIntegrationActionsProxy(sdkConfig)

	log.Printf("Deleting integration action draft %s", name)
	resp, err := iap.deleteIntegrationActionDraft(ctx, d.Id())
	if err != nil {
		if util.IsStatus404(resp) {
			log.Printf("Integration action draft already deleted %s", d.Id())
			return nil
		}
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete Integration action draft %s error: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := iap.getIntegrationActionDraftById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted Integration action draft %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error deleting integration action draft %s | error: %s", d.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("integration action draft %s still exists", d.Id()), resp))
	})
}

// Helper Functions

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
		Contract:      buildDraftContract(d),
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

func buildDraftContract(d *schema.ResourceData) *platformclientv2.Actioncontractinput {
	contractInput := d.Get("contract_input")
	//contractOutput := d.Get("contract_output").(string)

	var input platformclientv2.Postinputcontract
	var output platformclientv2.Postoutputcontract

	if contractInput != nil {
		// First check if it's a string
		if inputStr, ok := contractInput.(string); ok {
			// Convert string to map[string]interface{}
			var inputMap map[string]interface{}
			err := json.Unmarshal([]byte(inputStr), &inputMap)
			if err != nil {
				log.Printf("ERROR: failed to unmarshal contract_input: %v", err)
				return nil
			}
			input = *buildPostActionInput(inputMap)
		} else if inputMap, ok := contractInput.(map[string]interface{}); ok {
			input = *buildPostActionInput(inputMap)
		} else {
			log.Printf("ERROR: contract_input must be a string or map[string]interface{}")
			return nil
		}
	}

	log.Println(contractInput)
	log.Println(input)
	contract := &platformclientv2.Actioncontractinput{
		Input:  &input,
		Output: &output,
	}
	log.Println(contract.Input.String())
	return contract
}

func buildPostActionInput(input interface{}) *platformclientv2.Postinputcontract {
	if input == nil {
		return nil
	}

	inputSchema := &platformclientv2.Postinputcontract{}
	schema := &platformclientv2.Jsonschemadocument{}

	// Convert input to map[string]interface{}
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		log.Println("ERROR1")
		return nil
	}

	if schemaStr, ok := inputMap["$schema"].(string); ok {
		schema.Schema = &schemaStr
	}

	if title, ok := inputMap["title"].(string); ok {
		schema.Title = &title
	}

	if description, ok := inputMap["description"].(string); ok {
		schema.Description = &description
	}

	if typeStr, ok := inputMap["type"].(string); ok {
		schema.VarType = &typeStr
	}

	// Handle required array
	if required, ok := inputMap["required"].([]interface{}); ok {
		requiredStr := make([]string, len(required))
		for i, r := range required {
			requiredStr[i] = r.(string)
		}
		schema.Required = &requiredStr
	}

	// Handle properties map
	if inputMap["properties"] != "" {
		properties := inputMap["properties"].(map[string]interface{})
		schema.Properties = &properties
	}

	// Handle additionalProperties
	if additionalProps, ok := inputMap["additionalProperties"].(interface{}); ok {
		schema.AdditionalProperties = &additionalProps
	}

	inputSchema.InputSchema = schema
	return inputSchema
}

// flattenActionDraftContract converts the custom ActionContract into a JSON-encoded string
func flattenActionDraftContract(schema interface{}) (string, diag.Diagnostics) {
	if schema == nil {
		return "", nil
	}
	schemaBytes, err := json.Marshal(schema)
	if err != nil {
		return "", util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Error marshalling action contract %v", schema), err)
	}
	return string(schemaBytes), nil
}

// flattenActionDraftConfigRequest converts the platformclientv2.Requestconfig into a map
func FlattenActionDraftConfigRequest(sdkRequest platformclientv2.Requestconfig) []interface{} {
	requestMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(requestMap, "request_url_template", sdkRequest.RequestUrlTemplate)
	resourcedata.SetMapValueIfNotNil(requestMap, "request_type", sdkRequest.RequestType)
	resourcedata.SetMapValueIfNotNil(requestMap, "request_template", sdkRequest.RequestTemplate)
	resourcedata.SetMapValueIfNotNil(requestMap, "headers", sdkRequest.Headers)

	return []interface{}{requestMap}
}

// FlattenActionDraftConfigResponse converts the the platformclientv2.Responseconfig into a map
func FlattenActionDraftConfigResponse(sdkResponse platformclientv2.Responseconfig) []interface{} {
	responseMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(responseMap, "translation_map", sdkResponse.TranslationMap)
	resourcedata.SetMapValueIfNotNil(responseMap, "translation_map_defaults", sdkResponse.TranslationMapDefaults)
	resourcedata.SetMapValueIfNotNil(responseMap, "success_template", sdkResponse.SuccessTemplate)

	return []interface{}{responseMap}
}

func buildActionDraftFromResourceDataForUpdate(d *schema.ResourceData) *platformclientv2.Updatedraftinput {
	return &platformclientv2.Updatedraftinput{
		Name:     platformclientv2.String(d.Get("name").(string)),
		Category: platformclientv2.String(d.Get("category").(string)),
		Secure:   platformclientv2.Bool(d.Get("secure").(bool)),
		Contract: buildDraftContract(d),
		Config:   buildSdkActionConfig(d),
	}
}
