package integration

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The resource_genesyscloud_integration.go contains all of the methods that perform the core logic for a resource.
In general a resource should have a approximately 5 methods in it:

1.  A getAll.... function that the CX as Code exporter will use during the process of exporting Genesys Cloud.
2.  A create.... function that the resource will use to create a Genesys Cloud object (e.g. genesycloud_integration)
3.  A read.... function that looks up a single resource.
4.  An update... function that updates a single resource.
5.  A delete.... function that deletes a single resource.

Two things to note:

 1. All code in these methods should be focused on getting data in and out of Terraform.  All code that is used for interacting
    with a Genesys API should be encapsulated into a proxy class contained within the package.

 2. In general, to keep this file somewhat manageable, if you find yourself with a number of helper functions move them to a

utils function in the package.  This will keep the code manageable and easy to work through.
*/

// getAllIntegrations retrieves all of the integrations via Terraform in the Genesys Cloud and is used for the exporter
func getAllIntegrations(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	ip := getIntegrationsProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	integrations, resp, err := ip.getAllIntegrations(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to get all integrations %s", err), resp)
	}

	for _, integration := range *integrations {
		log.Printf("Dealing with integration id : %s", *integration.Id)
		resources[*integration.Id] = &resourceExporter.ResourceMeta{Name: *integration.Name}
	}
	return resources, nil
}

// createIntegration is used by the integrations resource to create Genesyscloud integration
func createIntegration(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	intendedState := d.Get("intended_state").(string)
	integrationType := d.Get("integration_type").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	ip := getIntegrationsProxy(sdkConfig)

	createIntegrationReq := &platformclientv2.Createintegrationrequest{
		IntegrationType: &platformclientv2.Integrationtype{
			Id: &integrationType,
		},
	}
	integration, resp, err := ip.createIntegration(ctx, createIntegrationReq)
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to create integration error: %s", err), resp)
	}

	d.SetId(*integration.Id)

	//Update integration config separately
	diagErr, name := updateIntegrationConfigFromResourceData(ctx, d, ip)
	if diagErr != nil {
		return diagErr
	}

	// Set attributes that can only be modified in a patch
	if d.HasChange("intended_state") {
		log.Printf("Updating additional attributes for integration %s", name)
		_, resp, patchErr := ip.updateIntegration(ctx, d.Id(), &platformclientv2.Integration{
			IntendedState: &intendedState,
		})
		if patchErr != nil {
			return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to update integration %s error: %s", d.Id(), err), resp)
		}
	}
	log.Printf("Created integration %s %s", name, *integration.Id)
	return readIntegration(ctx, d, meta)
}

// readIntegration is used by the integration resource to read an integration from genesys cloud.
func readIntegration(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	ip := getIntegrationsProxy(sdkConfig)

	log.Printf("Reading integration %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		currentIntegration, resp, getErr := ip.getIntegrationById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("failed to read integration %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("failed to read integration %s | error: %s", d.Id(), getErr), resp))
		}

		d.Set("integration_type", *currentIntegration.IntegrationType.Id)
		resourcedata.SetNillableValue(d, "intended_state", currentIntegration.IntendedState)

		// Use returned ID to get current config, which contains complete configuration
		integrationConfig, resp, err := ip.getIntegrationConfig(ctx, *currentIntegration.Id)

		if err != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("failed to read config of integration %s | error: %s", d.Id(), getErr), resp))
		}

		d.Set("config", flattenIntegrationConfig(integrationConfig))
		log.Printf("Read integration %s %s", d.Id(), *currentIntegration.Name)
		return nil
	})
}

// updateIntegration is used by the integration resource to update an integration in Genesys Cloud
func updateIntegration(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	intendedState := d.Get("intended_state").(string)
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	ip := getIntegrationsProxy(sdkConfig)

	diagErr, name := updateIntegrationConfigFromResourceData(ctx, d, ip)
	if diagErr != nil {
		return diagErr
	}

	if d.HasChange("intended_state") {
		log.Printf("Updating integration %s", name)
		_, resp, patchErr := ip.updateIntegration(ctx, d.Id(), &platformclientv2.Integration{
			IntendedState: &intendedState,
		})
		if patchErr != nil {
			return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to update Integration %s %s", d.Id(), patchErr), resp)
		}
	}
	log.Printf("Updated integration %s %s", name, d.Id())
	return readIntegration(ctx, d, meta)
}

// deleteIntegration is used by the integration resource to delete an integration from Genesys cloud.
func deleteIntegration(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	ip := getIntegrationsProxy(sdkConfig)

	resp, err := ip.deleteIntegration(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to delete Integration %s error: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := ip.getIntegrationById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				// Integration deleted
				log.Printf("Deleted Integration %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("error deleting integration %s | error: %s", d.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("integration %s still exists", d.Id()), resp))
	})
}
