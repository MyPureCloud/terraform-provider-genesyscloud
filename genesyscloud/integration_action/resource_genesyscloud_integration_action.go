package integration_action

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
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

// getAllIntegrationActions retrieves all of the integration action via Terraform in the Genesys Cloud and is used for the exporter
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
		resources[*action.Id] = &resourceExporter.ResourceMeta{BlockLabel: *action.Name}
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

// readIntegrationAction is used by the integration action resource to read an action from genesys cloud.
func readIntegrationAction(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	iap := getIntegrationActionsProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceIntegrationAction(), constants.ConsistencyChecks(), ResourceType)

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
			d.Set("config_request", FlattenActionConfigRequest(*action.Config.Request))
		} else {
			d.Set("config_request", nil)
		}

		if action.Config != nil && action.Config.Response != nil {
			action.Config.Response.SuccessTemplate = successTemp
			d.Set("config_response", FlattenActionConfigResponse(*action.Config.Response))
		} else {
			d.Set("config_response", nil)
		}

		log.Printf("Read integration action %s %s", d.Id(), *action.Name)
		return cc.CheckState(d)
	})
}

// updateIntegrationAction is used by the integration action resource to update an action in Genesys Cloud
func updateIntegrationAction(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	iap := getIntegrationActionsProxy(sdkConfig)

	name := d.Get("name").(string)
	category := d.Get("category").(string)

	log.Printf("Updating integration action %s", name)

	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get the latest action version to send with PATCH
		action, resp, err := iap.getIntegrationActionById(ctx, d.Id())
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
