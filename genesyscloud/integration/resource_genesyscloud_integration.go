package integration

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

func getAllIntegrations(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	ip := getIntegrationsProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	integrations, err := ip.getAllIntegrations(ctx)
	if err != nil {
		return nil, diag.Errorf("Failed to get all integrations: %v", err)
	}

	for _, integration := range *integrations {
		log.Printf("Dealing with integration id : %s", *integration.Id)
		resources[*integration.Id] = &resourceExporter.ResourceMeta{Name: *integration.Name}
	}

	return resources, nil
}

func createIntegration(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	intendedState := d.Get("intended_state").(string)
	integrationType := d.Get("integration_type").(string)

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	ip := getIntegrationsProxy(sdkConfig)

	createIntegrationReq := &platformclientv2.Createintegrationrequest{
		IntegrationType: &platformclientv2.Integrationtype{
			Id: &integrationType,
		},
	}
	integration, err := ip.createIntegration(ctx, createIntegrationReq)
	if err != nil {
		return diag.Errorf("Failed to create integration: %s", err)
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
		_, patchErr := ip.updateIntegration(ctx, d.Id(), &platformclientv2.Integration{
			IntendedState: &intendedState,
		})

		if patchErr != nil {
			return diag.Errorf("Failed to update integration %s: %v", name, patchErr)
		}
	}

	log.Printf("Created integration %s %s", name, *integration.Id)
	return readIntegration(ctx, d, meta)
}

func readIntegration(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	ip := getIntegrationsProxy(sdkConfig)

	log.Printf("Reading integration %s", d.Id())

	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		const pageSize = 100
		const pageNum = 1
		currentIntegration, resp, getErr := ip.getIntegrationById(ctx, d.Id())
		if getErr != nil {
			if gcloud.IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read integration %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read integration %s: %s", d.Id(), getErr))
		}

		d.Set("integration_type", *currentIntegration.IntegrationType.Id)
		if currentIntegration.IntendedState != nil {
			d.Set("intended_state", *currentIntegration.IntendedState)
		} else {
			d.Set("intended_state", nil)
		}

		// Use returned ID to get current config, which contains complete configuration
		integrationConfig, _, err := ip.getIntegrationConfig(ctx, *currentIntegration.Id)

		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("Failed to read config of integration %s: %s", d.Id(), getErr))
		}

		d.Set("config", flattenIntegrationConfig(integrationConfig))

		log.Printf("Read integration %s %s", d.Id(), *currentIntegration.Name)

		return nil
	})
}

func updateIntegration(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	intendedState := d.Get("intended_state").(string)

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	ip := getIntegrationsProxy(sdkConfig)

	diagErr, name := updateIntegrationConfigFromResourceData(ctx, d, ip)
	if diagErr != nil {
		return diagErr
	}

	if d.HasChange("intended_state") {
		log.Printf("Updating integration %s", name)
		_, patchErr := ip.updateIntegration(ctx, d.Id(), &platformclientv2.Integration{
			IntendedState: &intendedState,
		})
		if patchErr != nil {
			return diag.Errorf("Failed to update integration %s: %s", name, patchErr)
		}
	}

	log.Printf("Updated integration %s %s", name, d.Id())
	return readIntegration(ctx, d, meta)
}

func deleteIntegration(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	ip := getIntegrationsProxy(sdkConfig)

	_, err := ip.deleteIntegration(ctx, d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete the integration %s: %s", d.Id(), err)
	}

	return gcloud.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := ip.getIntegrationById(ctx, d.Id())
		if err != nil {
			if gcloud.IsStatus404(resp) {
				// Integration deleted
				log.Printf("Deleted Integration %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting integration %s: %s", d.Id(), err))
		}
		return retry.RetryableError(fmt.Errorf("Integration %s still exists", d.Id()))
	})
}
