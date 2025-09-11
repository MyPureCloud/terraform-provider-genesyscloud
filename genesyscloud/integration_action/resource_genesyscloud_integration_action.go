package integration_action

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

/*
The resource_genesyscloud_integration_action.go contains all of the methods that perform the core logic for a resource.
In general a resource should have a approximately 5 methods in it:

1.  A getAll.... function that the CX as Code exporter will use during the process of exporting Genesys Cloud.
2.  A create.... function that the resource will use to create a Genesys Cloud object (e.g. genesycloud_integration_action)
3.  A read.... function that looks up a single resource.
4.  An update... function that updates a single resource.
5.  A delete.... function that deletes a single resource.

Two things to note:

 1. All code in these methods should be focused on getting data in and out of Terraform.  All code that is used for interacting
    with a Genesys API should be encapsulated into a proxy class contained within the package.

 2. In general, to keep this file somewhat manageable, if you find yourself with a number of helper functions move them to a

utils function in the package.  This will keep the code manageable and easy to work through.
*/

// getAllIntegrationActions retrieves all integration actions via Terraform in the Genesys Cloud and is used for the exporter
func getAllIntegrationActions(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	iap := getIntegrationActionsProxy(clientConfig)

	actions, resp, err := iap.getAllIntegrationActions(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get integration actions %s", err), resp)
	}

	for _, action := range *actions {
		// Don't include "static" actions
		if strings.HasPrefix(*action.Id, "static") {
			continue
		}
		blockHash, err := util.QuickHashFields(action.Category)
		if err != nil {
			return nil, diag.Errorf("error hashing integration action %s: %s", *action.Name, err)
		}
		resources[*action.Id] = &resourceExporter.ResourceMeta{BlockLabel: *action.Name, BlockHash: blockHash}
	}
	return resources, nil
}

// createIntegrationAction is used by the integration actions resource to create Genesyscloud integration action
func createIntegrationAction(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	category := d.Get("category").(string)
	integrationId := d.Get("integration_id").(string)
	secure := d.Get("secure").(bool)
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	iap := getIntegrationActionsProxy(sdkConfig)

	if containsFunctionDataAction(category) {
		return createFunctionDataActionDraft(ctx, d, meta, iap)
	}

	log.Printf("Creating integration action %s", name)

	actionContract, diagErr := BuildSdkActionContract(d)
	if diagErr != nil {
		return diagErr
	}

	diagErr = util.RetryWhen(util.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		action, resp, err := iap.createIntegrationAction(ctx, &IntegrationAction{
			Name:          &name,
			Category:      &category,
			IntegrationId: &integrationId,
			Secure:        &secure,
			Contract:      actionContract,
			Config:        BuildSdkActionConfig(d),
		})
		if err != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create integration action %s error: %s", name, err), resp)
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

func containsFunctionDataAction(s string) bool {
	normalized := strings.ToLower(s)
	normalized = strings.ReplaceAll(normalized, "_", " ")
	return strings.Contains(normalized, "function data action")
}

func updateFunctionDataActionDraft(ctx context.Context, d *schema.ResourceData, meta interface{}, iap *integrationActionsProxy) diag.Diagnostics {
	id := d.Id()

	integrationId := d.Get("integration_id").(string)
	name := d.Get("name").(string)
	category := d.Get("category").(string)
	secure := d.Get("secure").(bool)

	version := 1
	zipid := ""

	// Get file_path from function_config
	var filePath string
	if functionConfig := d.Get("function_config"); functionConfig != nil {
		if configList := functionConfig.([]interface{}); len(configList) > 0 {
			if configMap := configList[0].(map[string]interface{}); configMap != nil {
				if filePathVal, exists := configMap["file_path"]; exists {
					filePath = filePathVal.(string)
				}
			}
		}
	}

	log.Printf("Updating integration action Function%s", name)

	actionContract, diagErr := BuildSdkActionContract(d)
	if diagErr != nil {
		return diagErr
	}

	diagErr = util.RetryWhen(util.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		action, resp, err := iap.createIntegrationActionDraft(ctx, &IntegrationAction{
			Name:          &name,
			Category:      &category,
			Id:            &id,
			IntegrationId: &integrationId,
			Secure:        &secure,
			Contract:      actionContract,
			Config:        BuildSdkActionConfig(d),
		})
		if err != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create integration action %s error: %s", name, err), resp)
		}
		d.SetId(*action.Id)
		id = *action.Id
		log.Printf("Created integration action %s %s", name, *action.Id)
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	diagErr = util.RetryWhen(util.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		resp, err := iap.uploadIntegrationActionDraftFunction(ctx, id, filePath)
		if err != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create integration action %s error: %s", name, err), resp)
		}
		log.Printf("Uploaded function zip for integration action %s %s", name, id)
		return resp, nil
	}, 501)
	if diagErr != nil {
		return diagErr
	}

	// get function for zip id
	diagErr = util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		functionData, _, err := iap.getIntegrationActionDraftFunction(ctx, id)
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("Failed to get function for integration action %s error: %s", name, err))
		}

		//zipid
		zipid, err = extractZipIdFromFunctionData(functionData)

		// use zipid in function settings
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("Failed to get zipId for integration action %s error: %s", name, err))
		}

		// Check if zipId is empty and retry if it is
		if zipid == "" {
			log.Printf("DEBUG: zipId is empty, retrying...")
			time.Sleep(2 * time.Second)
			return retry.RetryableError(fmt.Errorf("zipId is empty, retrying"))
		}

		log.Printf("DEBUG: Got zipId: %s", zipid)
		return nil
	})
	if diagErr != nil {
		return diagErr
	}
	// update draft with function settings
	diagErr = util.RetryWhen(util.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get function config from resource data
		functionConfig := BuildSdkFunctionConfig(d, zipid)
		if functionConfig != nil && functionConfig.Function != nil {
			_, resp, err := iap.updateIntegrationActionDraftWithFunction(ctx, id, functionConfig.Function)
			if err != nil {
				return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update integration action %s error: %s", name, err), resp)
			}
			// Note: Functionconfig doesn't have Version field, so we can't update version here
			return resp, nil
		}
		return nil, diag.Errorf("No function configuration found")
	})
	if diagErr != nil {
		return diagErr
	}

	// get latest version
	diagErr = util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get the latest action version to send with PATCH
		action, resp, err := iap.getIntegrationActionDraftById(ctx, d.Id())
		if err != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read integration action %s error: %s", d.Id(), err), resp)
		}

		version = *action.Version
		log.Printf("DEBUG: Got version from draft: %d", version)
		return resp, nil
	})

	log.Printf("DEBUG: Publishing action as publish=true")
	diagErr = util.RetryWhen(util.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("DEBUG: Updating Published draft with version: %d", version)
		resp, err := iap.publishIntegrationActionDraft(ctx, id, version+1)
		if err != nil {
			return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to publish integration action %s error: %s", name, err), resp)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return readIntegrationActionFunction(ctx, d, meta)
}

func createFunctionDataActionDraft(ctx context.Context, d *schema.ResourceData, meta interface{}, iap *integrationActionsProxy) diag.Diagnostics {
	category := d.Get("category").(string)
	id := ""
	version := 1
	zipid := ""
	name := d.Get("name").(string)
	integrationId := d.Get("integration_id").(string)
	secure := d.Get("secure").(bool)

	// Get file_path from function_config
	var filePath string
	if functionConfig := d.Get("function_config"); functionConfig != nil {
		if configList := functionConfig.([]interface{}); len(configList) > 0 {
			if configMap := configList[0].(map[string]interface{}); configMap != nil {
				if filePathVal, exists := configMap["file_path"]; exists {
					filePath = filePathVal.(string)
					log.Printf("DEBUG: file_path extracted from function_config: %s", filePath)
				}
			}
		}
	}

	if filePath == "" {
		log.Printf("DEBUG: file_path is empty, skipping function upload")
		return diag.Errorf("file_path is required in function_config for function data actions")
	}

	actionContract, diagErr := BuildSdkActionContract(d)
	if diagErr != nil {
		return diagErr
	}

	diagErr = util.RetryWhen(util.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		action, resp, err := iap.createIntegrationActionDraft(ctx, &IntegrationAction{
			Name:          &name,
			Category:      &category,
			IntegrationId: &integrationId,
			Secure:        &secure,
			Contract:      actionContract,
			Config:        BuildSdkActionConfig(d),
		})
		if err != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create integration action %s error: %s", name, err), resp)
		}
		d.SetId(*action.Id)
		id = *action.Id
		log.Printf("Created integration action %s %s", name, *action.Id)
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}
	// upload function zip
	diagErr = util.RetryWhen(util.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		resp, err := iap.uploadIntegrationActionDraftFunction(ctx, id, filePath)
		if err != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create integration action %s error: %s", name, err), resp)
		}
		log.Printf("Uploaded function zip for integration action %s %s", name, id)
		return resp, nil
	}, 501, 500)
	if diagErr != nil {
		return diagErr
	}

	// get function for zip id
	diagErr = util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		functionData, _, err := iap.getIntegrationActionDraftFunction(ctx, id)
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("Failed to get function for integration action %s error: %s", name, err))
		}

		//zipid
		zipid, err = extractZipIdFromFunctionData(functionData)

		// use zipid in function settings
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("Failed to get zipId for integration action %s error: %s", name, err))
		}

		// Check if zipId is empty and retry if it is
		if zipid == "" {
			log.Printf("DEBUG: zipId is empty, retrying...")
			time.Sleep(2 * time.Second)
			return retry.RetryableError(fmt.Errorf("zipId is empty, retrying"))
		}

		log.Printf("DEBUG: Got zipId: %s", zipid)
		return nil
	})
	if diagErr != nil {
		return diagErr
	}

	// update draft with function settings
	diagErr = util.RetryWhen(util.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get function config from resource data
		functionConfig := BuildSdkFunctionConfig(d, zipid)
		log.Printf("DEBUG: Built function config with zipId: %s", zipid)
		if functionConfig != nil && functionConfig.Function != nil {
			log.Printf("DEBUG: functionConfig.Function: %+v", functionConfig.Function)
			_, resp, err := iap.updateIntegrationActionDraftWithFunction(ctx, id, functionConfig.Function)
			if err != nil {
				return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update integration action %s error: %s", name, err), resp)
			}
			return resp, nil
		}
		return nil, diag.Errorf("No function configuration found")
	})
	if diagErr != nil {
		return diagErr
	}

	// get latest version
	diagErr = util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get the latest action version to send with PATCH
		action, resp, err := iap.getIntegrationActionDraftById(ctx, d.Id())
		if err != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read integration action %s error: %s", d.Id(), err), resp)
		}

		version = *action.Version
		log.Printf("DEBUG: Got version from draft: %d", version)
		return resp, nil
	})

	if diagErr != nil {
		return diagErr
	}

	log.Printf("DEBUG: Publishing action as publish=true")
	diagErr = util.RetryWhen(util.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("DEBUG: Publishing draft with version: %d", version)
		resp, err := iap.publishIntegrationActionDraft(ctx, id, version)
		if err != nil {
			return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to publish integration action %s error: %s", name, err), resp)
		}
		log.Printf("DEBUG: resp is not null and err is null...")
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return readIntegrationAction(ctx, d, meta)
}

// extractZipIdFromFunctionData extracts the zipId from the function data response
func extractZipIdFromFunctionData(functionData *platformclientv2.Functionconfig) (string, error) {
	if functionData == nil {
		return "", fmt.Errorf("function data is nil")
	}

	// Check if function has zipId
	if functionData.Function != nil && functionData.Function.ZipId != nil {
		return *functionData.Function.ZipId, nil
	}

	// Fallback to zip.id if function.zipId is not available
	if functionData.Zip != nil && functionData.Zip.Id != nil {
		return *functionData.Zip.Id, nil
	}

	return "", fmt.Errorf("zipId not found in function data")
}

// readIntegrationActionFunction is used by the integration action resource to read an action from genesys cloud.
func readIntegrationActionFunction(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	iap := getIntegrationActionsProxy(sdkConfig)

	log.Printf("Reading integration action function %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		var action *platformclientv2.Action
		var resp *platformclientv2.APIResponse
		var err error

		log.Printf("DEBUG: Reading published version of integration action %s", d.Id())
		action, resp, err = iap.integrationsApi.GetIntegrationsAction(d.Id(), "", true, true)

		if err != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read integration action %s | error: %s", d.Id(), err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read integration action %s | error: %s", d.Id(), err), resp))
		}

		// Retrieve config request/response templates
		reqTemp, resp, err := iap.getIntegrationActionTemplate(ctx, d.Id(), reqTemplateFileName)
		if err != nil {
			if util.IsStatus404(resp) {
				d.SetId("")
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read request template for integration action %s | error: %s", d.Id(), err), resp))
		}

		successTemp, resp, err := iap.getIntegrationActionTemplate(ctx, d.Id(), successTemplateFileName)
		if err != nil {
			if util.IsStatus404(resp) {
				d.SetId("")
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read success template for integration action %s | error: %s", d.Id(), err), resp))
		}

		resourcedata.SetNillableValue(d, "name", action.Name)
		resourcedata.SetNillableValue(d, "category", action.Category)
		resourcedata.SetNillableValue(d, "integration_id", action.IntegrationId)
		resourcedata.SetNillableValue(d, "secure", action.Secure)
		resourcedata.SetNillableValue(d, "config_timeout_seconds", action.Config.TimeoutSeconds)

		if action.Contract != nil && action.Contract.Input != nil && action.Contract.Input.InputSchema != nil {
			input, err := flattenActionContract(*action.Contract.Input.InputSchema)
			if err != nil {
				return retry.NonRetryableError(fmt.Errorf("%v", err))
			}
			_ = d.Set("contract_input", input)
		} else {
			_ = d.Set("contract_input", nil)
		}

		if action.Contract != nil && action.Contract.Output != nil && action.Contract.Output.SuccessSchema != nil {
			output, err := flattenActionContract(*action.Contract.Output.SuccessSchema)
			if err != nil {
				return retry.NonRetryableError(fmt.Errorf("%v", err))
			}
			_ = d.Set("contract_output", output)
		} else {
			_ = d.Set("contract_output", nil)
		}

		if action.Config != nil && action.Config.Request != nil {
			action.Config.Request.RequestTemplate = reqTemp
			_ = d.Set("config_request", FlattenActionConfigRequest(*action.Config.Request))
		} else {
			_ = d.Set("config_request", nil)
		}

		if action.Config != nil && action.Config.Response != nil {
			action.Config.Response.SuccessTemplate = successTemp
			_ = d.Set("config_response", FlattenActionConfigResponse(*action.Config.Response))
		} else {
			_ = d.Set("config_response", nil)
		}

		var functionData *platformclientv2.Functionconfig

		log.Printf("DEBUG: Reading published function for integration action %s", d.Id())
		functionData, resp, err = iap.getIntegrationActionFunction(ctx, d.Id())
		if err != nil {
			log.Printf("DEBUG: Could not read published function, skipping function data")
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read integration action %s | error: %s", d.Id(), err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read integration action %s | error: %s", d.Id(), err), resp))

		}

		if functionData != nil {
			action.Config.Request.RequestTemplate = reqTemp
			_ = d.Set("function_config", FlattenFunctionConfigRequest(*functionData))
		} else {
			_ = d.Set("function_config", nil)
		}

		log.Printf("Read integration action %s %s", d.Id(), *action.Name)
		return nil
	})
}

// readIntegrationAction is used by the integration action resource to read an action from genesys cloud.
func readIntegrationAction(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	iap := getIntegrationActionsProxy(sdkConfig)
	//cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceIntegrationAction(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading integration action %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		action, resp, err := iap.getIntegrationActionById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read integration action %s | error: %s", d.Id(), err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read integration action %s | error: %s", d.Id(), err), resp))
		}

		// Retrieve config request/response templates
		reqTemp, resp, err := iap.getIntegrationActionTemplate(ctx, d.Id(), reqTemplateFileName)
		if err != nil {
			if util.IsStatus404(resp) {
				d.SetId("")
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read request template for integration action %s | error: %s", d.Id(), err), resp))
		}

		successTemp, resp, err := iap.getIntegrationActionTemplate(ctx, d.Id(), successTemplateFileName)
		if err != nil {
			if util.IsStatus404(resp) {
				d.SetId("")
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read success template for integration action %s | error: %s", d.Id(), err), resp))
		}

		resourcedata.SetNillableValue(d, "name", action.Name)
		resourcedata.SetNillableValue(d, "category", action.Category)
		resourcedata.SetNillableValue(d, "integration_id", action.IntegrationId)
		resourcedata.SetNillableValue(d, "secure", action.Secure)
		resourcedata.SetNillableValue(d, "config_timeout_seconds", action.Config.TimeoutSeconds)

		if action.Contract != nil && action.Contract.Input != nil && action.Contract.Input.InputSchema != nil {
			input, err := flattenActionContract(*action.Contract.Input.InputSchema)
			if err != nil {
				return retry.NonRetryableError(fmt.Errorf("%v", err))
			}
			_ = d.Set("contract_input", input)
		} else {
			_ = d.Set("contract_input", nil)
		}

		if action.Contract != nil && action.Contract.Output != nil && action.Contract.Output.SuccessSchema != nil {
			output, err := flattenActionContract(*action.Contract.Output.SuccessSchema)
			if err != nil {
				return retry.NonRetryableError(fmt.Errorf("%v", err))
			}
			_ = d.Set("contract_output", output)
		} else {
			_ = d.Set("contract_output", nil)
		}

		if action.Config != nil && action.Config.Request != nil {
			action.Config.Request.RequestTemplate = reqTemp
			_ = d.Set("config_request", FlattenActionConfigRequest(*action.Config.Request))
		} else {
			_ = d.Set("config_request", nil)
		}

		if action.Config != nil && action.Config.Response != nil {
			action.Config.Response.SuccessTemplate = successTemp
			_ = d.Set("config_response", FlattenActionConfigResponse(*action.Config.Response))
		} else {
			_ = d.Set("config_response", nil)
		}

		if containsFunctionDataAction(*action.Category) {
			var functionData *platformclientv2.Functionconfig

			log.Printf("DEBUG: Reading published function for integration action %s", d.Id())
			functionData, resp, err = iap.getIntegrationActionFunction(ctx, d.Id())
			if err != nil {
				log.Printf("DEBUG: Could not read published function, skipping function data")
				if util.IsStatus404(resp) {
					return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read integration action %s | error: %s", d.Id(), err), resp))
				}
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read integration action %s | error: %s", d.Id(), err), resp))

			}

			if functionData != nil {
				action.Config.Request.RequestTemplate = reqTemp
				_ = d.Set("function_config", FlattenFunctionConfigRequest(*functionData))
			} else {
				_ = d.Set("function_config", nil)
			}
		}

		log.Printf("Read integration action %s %s", d.Id(), *action.Name)

		return nil
	})
}

// updateIntegrationAction is used by the integration action resource to update an action in Genesys Cloud
func updateIntegrationAction(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	iap := getIntegrationActionsProxy(sdkConfig)
	name := d.Get("name").(string)
	category := d.Get("category").(string)
	id := d.Id()

	log.Printf("Updating integration action %s", name)

	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get the latest action version to send with PATCH
		action, resp, err := iap.getIntegrationActionById(ctx, id)
		if err != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read integration action %s error: %s", d.Id(), err), resp)
		}

		_, resp, err = iap.updateIntegrationAction(ctx, d.Id(), &platformclientv2.Updateactioninput{
			Name:     &name,
			Category: &category,
			Version:  action.Version,
			Config:   BuildSdkActionConfig(d),
		})
		if err != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update integration action %s error: %s", name, err), resp)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	if containsFunctionDataAction(category) {
		return updateFunctionDataActionDraft(ctx, d, meta, iap)
	}

	log.Printf("Updated integration action %s", name)
	return readIntegrationAction(ctx, d, meta)
}

// deleteIntegrationAction is used by the integration action resource to delete an action from Genesys cloud.
func deleteIntegrationAction(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	iap := getIntegrationActionsProxy(sdkConfig)

	log.Printf("Deleting integration action %s", name)
	resp, err := iap.deleteIntegrationAction(ctx, d.Id())
	if err != nil {
		if util.IsStatus404(resp) {
			// Parent integration was probably deleted which caused the action to be deleted
			log.Printf("Integration action already deleted %s", d.Id())
			return nil
		}
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete Integration action %s error: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := iap.getIntegrationActionById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				// Integration action deleted
				log.Printf("Deleted Integration action %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error deleting integration action %s | error: %s", d.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("integration action %s still exists", d.Id()), resp))
	})
}
