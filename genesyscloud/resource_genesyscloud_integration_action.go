package genesyscloud

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

var (
	actionConfigRequest = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"request_url_template": {
				Description: "URL that may include placeholders for requests to 3rd party service.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"request_type": {
				Description:  "HTTP method to use for request (GET | PUT | POST | PATCH | DELETE).",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"GET", "PUT", "POST", "PATCH", "DELETE"}, false),
			},
			"request_template": {
				Description: "Velocity template to define request body sent to 3rd party service. Any instances of '${' must be properly escaped as '$${'",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"headers": {
				Description: "Map of headers in name, value pairs to include in request.",
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}

	actionConfigResponse = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"translation_map": {
				Description: "Map 'attribute name' and 'JSON path' pairs used to extract data from REST response.",
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"translation_map_defaults": {
				Description: "Map 'attribute name' and 'default value' pairs used as fallback values if JSON path extraction fails for specified key.",
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"success_template": {
				Description: "Velocity template to build response to return from Action. Any instances of '${' must be properly escaped as '$${'.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
		},
	}
)

func getAllIntegrationActions(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	integAPI := platformclientv2.NewIntegrationsApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		actions, _, getErr := integAPI.GetIntegrationsActions(pageSize, pageNum, "", "", "", "", "", "", "", "", "")
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of integration actions: %v", getErr)
		}

		if actions.Entities == nil || len(*actions.Entities) == 0 {
			break
		}

		for _, action := range *actions.Entities {
			// Don't include "static" actions
			if strings.HasPrefix(*action.Id, "static") {
				continue
			}
			resources[*action.Id] = &resourceExporter.ResourceMeta{Name: *action.Name}
		}
	}

	return resources, nil
}

func IntegrationActionExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: GetAllWithPooledClient(getAllIntegrationActions),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"integration_id": {RefType: "genesyscloud_integration"},
		},
		JsonEncodeAttributes: []string{"contract_input", "contract_output"},
	}
}

func ResourceIntegrationAction() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Integration Actions. See this page for detailed information on configuring Actions: https://help.mypurecloud.com/articles/add-configuration-custom-actions-integrations/",

		CreateContext: CreateWithPooledClient(createIntegrationAction),
		ReadContext:   ReadWithPooledClient(readIntegrationAction),
		UpdateContext: UpdateWithPooledClient(updateIntegrationAction),
		DeleteContext: DeleteWithPooledClient(deleteIntegrationAction),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description:  "Name of the action. Can be up to 256 characters long",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
			"category": {
				Description:  "Category of action. Can be up to 256 characters long.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
			"integration_id": {
				Description: "The ID of the integration this action is associated with. Changing the integration_id attribute will cause the existing integration_action to be dropped and recreated with a new ID.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"secure": {
				Description: "Indication of whether or not the action is designed to accept sensitive data. Changing the secure attribute will cause the existing integration_action to be dropped and recreated with a new ID.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
			},
			"config_timeout_seconds": {
				Description:  "Optional 1-60 second timeout enforced on the execution or test of this action. This setting is invalid for Custom Authentication Actions.",
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 60),
			},
			"contract_input": {
				Description:      "JSON Schema that defines the body of the request that the client (edge/architect/postman) is sending to the service, on the /execute path. Changing the contract_input attribute will cause the existing integration_action to be dropped and recreated with a new ID.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: suppressEquivalentJsonDiffs,
			},
			"contract_output": {
				Description:      "JSON schema that defines the transformed, successful result that will be sent back to the caller. Changing the contract_output attribute will cause the existing integration_action to be dropped and recreated with a new ID.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: suppressEquivalentJsonDiffs,
			},
			"config_request": {
				Description: "Configuration of outbound request.",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem:        actionConfigRequest,
			},
			"config_response": {
				Description: "Configuration of response processing.",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Elem:        actionConfigResponse,
			},
		},
	}
}

func createIntegrationAction(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	category := d.Get("category").(string)
	integrationId := d.Get("integration_id").(string)
	secure := d.Get("secure").(bool)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	integAPI := platformclientv2.NewIntegrationsApiWithConfig(sdkConfig)

	log.Printf("Creating integration action %s", name)

	actionContract, diagErr := buildSdkActionContract(d)
	if diagErr != nil {
		return diagErr
	}

	diagErr = RetryWhen(IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		action, resp, err := sdkPostIntegrationAction(&IntegrationAction{
			Name:          &name,
			Category:      &category,
			IntegrationId: &integrationId,
			Secure:        &secure,
			Contract:      actionContract,
			Config:        buildSdkActionConfig(d),
		}, integAPI)
		if err != nil {
			return resp, diag.Errorf("Failed to create integration action %s: %s", name, err)
		}
		d.SetId(*action.Id)

		log.Printf("Created integration action %s %s", name, *action.Id)
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return readIntegrationAction(ctx, d, meta)
}

func readIntegrationAction(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	integAPI := platformclientv2.NewIntegrationsApiWithConfig(sdkConfig)

	log.Printf("Reading integration action %s", d.Id())

	return WithRetriesForRead(ctx, d, func() *retry.RetryError {
		action, resp, getErr := sdkGetIntegrationAction(d.Id(), integAPI)
		if getErr != nil {
			if IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read integration action %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read integration action %s: %s", d.Id(), getErr))
		}

		// Retrieve config request/response templates
		reqTemp, resp, getErr := sdkGetIntegrationActionTemplate(d.Id(), "requesttemplate.vm", integAPI)
		if getErr != nil {
			if IsStatus404(resp) {
				d.SetId("")
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read request template for integration action %s: %s", d.Id(), getErr))
		}

		successTemp, resp, getErr := sdkGetIntegrationActionTemplate(d.Id(), "successtemplate.vm", integAPI)
		if getErr != nil {
			if IsStatus404(resp) {
				d.SetId("")
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read success template for integration action %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceIntegrationAction())
		if action.Name != nil {
			d.Set("name", *action.Name)
		} else {
			d.Set("name", nil)
		}

		if action.Category != nil {
			d.Set("category", *action.Category)
		} else {
			d.Set("category", nil)
		}

		if action.IntegrationId != nil {
			d.Set("integration_id", *action.IntegrationId)
		} else {
			d.Set("integration_id", nil)
		}

		if action.Secure != nil {
			d.Set("secure", *action.Secure)
		} else {
			d.Set("secure", nil)
		}

		if action.Config.TimeoutSeconds != nil {
			d.Set("config_timeout_seconds", *action.Config.TimeoutSeconds)
		} else {
			d.Set("config_timeout_seconds", nil)
		}

		if action.Contract != nil && action.Contract.Input != nil && action.Contract.Input.InputSchema != nil {
			input, err := flattenActionContract(*action.Contract.Input.InputSchema)
			if err != nil {
				return retry.NonRetryableError(fmt.Errorf("%v", err))
			}
			d.Set("contract_input", input)
		} else {
			d.Set("contract_input", nil)
		}

		if action.Contract != nil && action.Contract.Output != nil && action.Contract.Output.SuccessSchema != nil {
			output, err := flattenActionContract(*action.Contract.Output.SuccessSchema)
			if err != nil {
				return retry.NonRetryableError(fmt.Errorf("%v", err))
			}
			d.Set("contract_output", output)
		} else {
			d.Set("contract_output", nil)
		}

		if action.Config != nil && action.Config.Request != nil {
			action.Config.Request.RequestTemplate = reqTemp
			d.Set("config_request", flattenActionConfigRequest(*action.Config.Request))
		} else {
			d.Set("config_request", nil)
		}

		if action.Config != nil && action.Config.Response != nil {
			action.Config.Response.SuccessTemplate = successTemp
			d.Set("config_response", flattenActionConfigResponse(*action.Config.Response))
		} else {
			d.Set("config_response", nil)
		}

		log.Printf("Read integration action %s %s", d.Id(), *action.Name)
		return cc.CheckState()
	})
}

func updateIntegrationAction(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	category := d.Get("category").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	integAPI := platformclientv2.NewIntegrationsApiWithConfig(sdkConfig)

	log.Printf("Updating integration action %s", name)

	diagErr := RetryWhen(IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get the latest action version to send with PATCH
		action, resp, getErr := sdkGetIntegrationAction(d.Id(), integAPI)
		if getErr != nil {
			return resp, diag.Errorf("Failed to read integration action %s: %s", d.Id(), getErr)
		}

		_, _, err := integAPI.PatchIntegrationsAction(d.Id(), platformclientv2.Updateactioninput{
			Name:     &name,
			Category: &category,
			Version:  action.Version,
			Config:   buildSdkActionConfig(d),
		})
		if err != nil {
			return resp, diag.Errorf("Failed to update integration action %s: %s", name, err)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated integration action %s", name)
	return readIntegrationAction(ctx, d, meta)
}

func deleteIntegrationAction(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	integAPI := platformclientv2.NewIntegrationsApiWithConfig(sdkConfig)

	log.Printf("Deleting integration action %s", name)
	resp, err := integAPI.DeleteIntegrationsAction(d.Id())
	if err != nil {
		if IsStatus404(resp) {
			// Parent integration was probably deleted which caused the action to be deleted
			log.Printf("Integration action already deleted %s", d.Id())
			return nil
		}
		return diag.Errorf("Failed to delete Integration action %s: %s", d.Id(), err)
	}

	return WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := sdkGetIntegrationAction(d.Id(), integAPI)
		if err != nil {
			if IsStatus404(resp) {
				// Integration action deleted
				log.Printf("Deleted Integration action %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting integration action %s: %s", d.Id(), err))
		}
		return retry.RetryableError(fmt.Errorf("Integration action %s still exists", d.Id()))
	})
}

func buildSdkActionContract(d *schema.ResourceData) (*ActionContract, diag.Diagnostics) {
	configInput := d.Get("contract_input").(string)
	inputVal, err := jsonStringToInterface(configInput)
	if err != nil {
		return nil, diag.Errorf("Failed to parse contract input %s: %v", configInput, err)
	}

	configOutput := d.Get("contract_output").(string)
	outputVal, err := jsonStringToInterface(configOutput)
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

// Overriding the SDK Action contract as it does not allow some JSON schema fields to be set such as 'items' for an array
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

func sdkPostIntegrationAction(body *IntegrationAction, api *platformclientv2.IntegrationsApi) (*IntegrationAction, *platformclientv2.APIResponse, error) {
	apiClient := &api.Configuration.APIClient

	// create path and map variables
	path := api.Configuration.BasePath + "/api/v2/integrations/actions"

	headerParams := make(map[string]string)

	// add default headers if any
	for key := range api.Configuration.DefaultHeader {
		headerParams[key] = api.Configuration.DefaultHeader[key]
	}

	headerParams["Authorization"] = "Bearer " + api.Configuration.AccessToken
	headerParams["Content-Type"] = "application/json"
	headerParams["Accept"] = "application/json"

	var successPayload *IntegrationAction
	response, err := apiClient.CallAPI(path, http.MethodPost, body, headerParams, nil, nil, "", nil)
	if err != nil {
		// Nothing special to do here, but do avoid processing the response
	} else if response.Error != nil {
		err = errors.New(response.ErrorMessage)
	} else {
		err = json.Unmarshal([]byte(response.RawBody), &successPayload)
	}
	return successPayload, response, err
}

func sdkGetIntegrationAction(actionId string, api *platformclientv2.IntegrationsApi) (*IntegrationAction, *platformclientv2.APIResponse, error) {
	apiClient := &api.Configuration.APIClient

	// create path and map variables
	path := api.Configuration.BasePath + "/api/v2/integrations/actions/" + actionId

	headerParams := make(map[string]string)
	queryParams := make(map[string]string)

	// oauth required
	if api.Configuration.AccessToken != "" {
		headerParams["Authorization"] = "Bearer " + api.Configuration.AccessToken
	}
	// add default headers if any
	for key := range api.Configuration.DefaultHeader {
		headerParams[key] = api.Configuration.DefaultHeader[key]
	}

	queryParams["expand"] = "contract"
	queryParams["includeConfig"] = "true"

	headerParams["Content-Type"] = "application/json"
	headerParams["Accept"] = "application/json"

	var successPayload *IntegrationAction
	response, err := apiClient.CallAPI(path, http.MethodGet, nil, headerParams, queryParams, nil, "", nil)
	if err != nil {
		// Nothing special to do here, but do avoid processing the response
	} else if response.Error != nil {
		err = errors.New(response.ErrorMessage)
	} else {
		err = json.Unmarshal([]byte(response.RawBody), &successPayload)
	}
	return successPayload, response, err
}

func sdkGetIntegrationActionTemplate(actionId, templateName string, api *platformclientv2.IntegrationsApi) (*string, *platformclientv2.APIResponse, error) {
	apiClient := &api.Configuration.APIClient

	// create path and map variables
	path := api.Configuration.BasePath + "/api/v2/integrations/actions/" + actionId + "/templates/" + templateName

	headerParams := make(map[string]string)
	queryParams := make(map[string]string)

	// oauth required
	if api.Configuration.AccessToken != "" {
		headerParams["Authorization"] = "Bearer " + api.Configuration.AccessToken
	}
	// add default headers if any
	for key := range api.Configuration.DefaultHeader {
		headerParams[key] = api.Configuration.DefaultHeader[key]
	}

	headerParams["Content-Type"] = "application/json"
	headerParams["Accept"] = "*/*"

	var successPayload *string
	response, err := apiClient.CallAPI(path, http.MethodGet, nil, headerParams, queryParams, nil, "", nil)
	if err != nil {
		// Nothing special to do here, but do avoid processing the response
	} else if response.Error != nil {
		err = errors.New(response.ErrorMessage)
	} else {
		templateStr := string(response.RawBody)
		successPayload = &templateStr
	}
	return successPayload, response, err
}
