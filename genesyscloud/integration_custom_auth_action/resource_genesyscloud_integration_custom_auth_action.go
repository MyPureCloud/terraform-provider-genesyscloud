package integration_custom_auth_action

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	integrationAction "terraform-provider-genesyscloud/genesyscloud/integration_action"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

/*
The resource_genesyscloud_integration_custom_auth_action.go contains all of the methods that perform the core logic for a resource.
In general a resource should have a approximately 5 methods in it:

1.  A getAll.... function that the CX as Code exporter will use during the process of exporting Genesys Cloud.
2.  A create.... function that the resource will use to create a Genesys Cloud object (e.g. genesyscloud_integration_custom_auth_action)
3.  A read.... function that looks up a single resource.
4.  An update... function that updates a single resource.
5.  A delete.... function that deletes a single resource.

Two things to note:

 1. All code in these methods should be focused on getting data in and out of Terraform.  All code that is used for interacting
    with a Genesys API should be encapsulated into a proxy class contained within the package.

 2. In general, to keep this file somewhat manageable, if you find yourself with a number of helper functions move them to a

utils function in the package.  This will keep the code manageable and easy to work through.
*/

// getAllModifiedCustomAuthActions retrieves only the custom auth actions that were modified at least
// once for use in the exporter (version > 1). ie. Unmodified custom auth actions are not to be exported since the defaults
// are created and managed by Genesys itself based on the Integration configuration.
func getAllModifiedCustomAuthActions(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	cap := getCustomAuthActionsProxy(clientConfig)

	actions, resp, err := cap.getAllIntegrationCustomAuthActions(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get integration custom auth actions error: %s", err), resp)
	}

	for _, action := range *actions {
		if *action.Version == 1 {
			continue
		}
		resources[*action.Id] = &resourceExporter.ResourceMeta{BlockLabel: *action.Name}
	}
	return resources, nil
}

// createIntegrationCustomAuthAction is used by the custom auth actions resource to manage the Genesyscloud integration custom auth action
func createIntegrationCustomAuthAction(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	cap := getCustomAuthActionsProxy(sdkConfig)

	integrationId := d.Get("integration_id").(string)
	authActionId := getCustomAuthIdFromIntegration(integrationId)

	name := resourcedata.GetNillableValue[string](d, "name")

	// Precheck that integration type and its credential type if it should have a custom auth data action
	if ok, err := isIntegrationAndCredTypesCorrect(ctx, cap, integrationId); !ok || err != nil {
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("configuration of integration %s does not allow for a custom auth data action", integrationId), err)
	}

	log.Printf("Retrieving the custom auth action of integration %s", integrationId)

	// Retrieve the automatically-generated custom auth action
	// to make sure it exists before updating
	diagErr := util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		authAction, resp, err := cap.getCustomAuthActionById(ctx, authActionId)
		if err != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("cannot find custom auth action of integration %s | error: %v", integrationId, err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error getting custom auth action %s | error: %s", d.Id(), err), resp))
		}

		// Get default name if not to be overriden
		if name == nil {
			name = authAction.Name
		}

		d.SetId(*authAction.Id)

		return nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updating custom auth action of integration %s", integrationId)

	// Update the custom auth action with the actual configuration
	diagErr = util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get the latest action version to send with PATCH
		action, resp, err := cap.getCustomAuthActionById(ctx, authActionId)
		if err != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read integration custom auth action %s error: %s", authActionId, err), resp)
		}

		_, resp, err = cap.updateCustomAuthAction(ctx, authActionId, &platformclientv2.Updateactioninput{
			Name:    name,
			Version: action.Version,
			Config:  BuildSdkCustomAuthActionConfig(d),
		})
		if err != nil {
			return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update integration action %s error: %s", *name, err), resp)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated custom auth action %s", *name)
	return readIntegrationCustomAuthAction(ctx, d, meta)
}

// readIntegrationCustomAuthAction is used by the integration action resource to read a custom auth action from genesys cloud
func readIntegrationCustomAuthAction(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	cap := getCustomAuthActionsProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceIntegrationCustomAuthAction(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading integration action %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		action, resp, err := cap.getCustomAuthActionById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read integration custom auth action %s | error: %s", d.Id(), err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read integration custom auth action %s | error: %s", d.Id(), err), resp))
		}

		// Retrieve config request/response templates
		reqTemp, resp, err := cap.getIntegrationActionTemplate(ctx, d.Id(), reqTemplateFileName)
		if err != nil {
			if util.IsStatus404(resp) {
				d.SetId("")
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read request template for integration action %s | error: %s", d.Id(), err), resp))
		}

		successTemp, resp, err := cap.getIntegrationActionTemplate(ctx, d.Id(), successTemplateFileName)
		if err != nil {
			if util.IsStatus404(resp) {
				d.SetId("")
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read success template for integration action %s | error: %s", d.Id(), err), resp))
		}

		resourcedata.SetNillableValue(d, "name", action.Name)
		resourcedata.SetNillableValue(d, "integration_id", action.IntegrationId)

		if action.Config != nil && action.Config.Request != nil {
			action.Config.Request.RequestTemplate = reqTemp
			d.Set("config_request", integrationAction.FlattenActionConfigRequest(*action.Config.Request))
		} else {
			d.Set("config_request", nil)
		}

		if action.Config != nil && action.Config.Response != nil {
			action.Config.Response.SuccessTemplate = successTemp
			d.Set("config_response", integrationAction.FlattenActionConfigResponse(*action.Config.Response))
		} else {
			d.Set("config_response", nil)
		}

		log.Printf("Read integration action %s %s", d.Id(), *action.Name)
		return cc.CheckState(d)
	})
}

// updateIntegrationCustomAuthAction is used by the integration action resource to update a custom auth in Genesys Cloud
func updateIntegrationCustomAuthAction(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	cap := getCustomAuthActionsProxy(sdkConfig)

	name := resourcedata.GetNillableValue[string](d, "name")

	log.Printf("Updating integration custom auth action %s", *name)

	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get the latest action version to send with PATCH
		action, resp, err := cap.getCustomAuthActionById(ctx, d.Id())
		if err != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read integration custom auth action %s error: %s", d.Id(), err), resp)
		}
		if name == nil {
			name = action.Name
		}

		_, resp, err = cap.updateCustomAuthAction(ctx, d.Id(), &platformclientv2.Updateactioninput{
			Name:    name,
			Version: action.Version,
			Config:  BuildSdkCustomAuthActionConfig(d),
		})
		if err != nil {
			return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update integration action %s error: %s", *name, err), resp)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated custom auth action %s", *name)
	return readIntegrationCustomAuthAction(ctx, d, meta)
}

// deleteIntegrationCustomAuthAction does not do anything as deleting a custom auth action is not possible
func deleteIntegrationCustomAuthAction(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	log.Printf("Removing terraform resource integration_custom_auth_action %s will not remove the Data Action itself in the org", name)
	log.Printf("The Custom Auth Data Action cannot be removed unless the Web Services Data Action Integration itself is deleted or if the Credentials type is changed from 'User Defined (OAuth)' to a different type")
	return nil
}
