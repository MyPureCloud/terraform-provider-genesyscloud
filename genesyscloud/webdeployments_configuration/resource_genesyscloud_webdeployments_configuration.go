package webdeployments_configuration

import (
	"context"
	"fmt"
	"log"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"
)

func getAllWebDeploymentConfigurations(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	wp := getWebDeploymentConfigurationsProxy(clientConfig)

	configurations, err := wp.getWebDeploymentsConfiguration(ctx)
	if err != nil {
		return nil, diag.Errorf("%v", err)
	}

	for _, configuration := range *configurations.Entities {
		resources[*configuration.Id] = &resourceExporter.ResourceMeta{Name: *configuration.Name}
	}

	return resources, nil
}

func waitForConfigurationDraftToBeActive(ctx context.Context, meta interface{}, id string) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	wp := getWebDeploymentConfigurationsProxy(sdkConfig)

	return gcloud.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		configuration, resp, err := wp.getWebdeploymentsConfigurationVersionsDraft(ctx, id)
		if err != nil {
			if gcloud.IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Error verifying active status for new web deployment configuration %s: %s", id, err))
			}
			return retry.NonRetryableError(fmt.Errorf("Error verifying active status for new web deployment configuration %s: %s", id, err))
		}

		if *configuration.Status == "Active" {
			return nil
		}

		return retry.RetryableError(fmt.Errorf("Web deployment configuration %s not active yet. Status: %s", id, *configuration.Status))
	})
}

func createWebDeploymentConfiguration(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	wp := getWebDeploymentConfigurationsProxy(sdkConfig)

	name, inputCfg := readWebDeploymentConfigurationFromResourceData(d)
	log.Printf("Creating web deployment configuration %s", name)

	diagErr := gcloud.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		configuration, resp, err := wp.createWebdeploymentsConfiguration(ctx, *inputCfg)
		if err != nil {
			var extraErrorInfo string
			featureIsNotImplemented, fieldName := featureNotImplemented(resp)
			if featureIsNotImplemented {
				extraErrorInfo = fmt.Sprintf("Feature '%s' is not yet implemented", fieldName)
			}
			if gcloud.IsStatus400(resp) {
				return retry.RetryableError(fmt.Errorf("failed to create web deployment configuration %s: %s. %s", name, err, extraErrorInfo))
			}
			return retry.NonRetryableError(fmt.Errorf("failed to create web deployment configuration %s: %s. %s", name, err, extraErrorInfo))
		}
		d.SetId(*configuration.Id)
		d.Set("status", configuration.Status)

		return nil
	})
	if diagErr != nil {
		return diagErr
	}

	activeError := waitForConfigurationDraftToBeActive(ctx, meta, d.Id())
	if activeError != nil {
		return diag.Errorf("Web deployment configuration %s did not become active and could not be published", name)
	}

	diagErr = gcloud.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		configuration, resp, err := wp.createWebdeploymentsConfigurationVersionsDraftPublish(ctx, d.Id())
		if err != nil {
			if gcloud.IsStatus400(resp) {
				return retry.RetryableError(fmt.Errorf("Error publishing web deployment configuration %s: %s", name, err))
			}
			return retry.NonRetryableError(fmt.Errorf("Error publishing web deployment configuration %s: %s", name, err))
		}
		d.Set("version", configuration.Version)
		d.Set("status", configuration.Status)
		log.Printf("Created web deployment configuration %s %s", name, *configuration.Id)

		return nil
	})
	if diagErr != nil {
		return diagErr
	}

	return readWebDeploymentConfiguration(ctx, d, meta)
}

func readWebDeploymentConfiguration(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	wp := getWebDeploymentConfigurationsProxy(sdkConfig)

	version := d.Get("version").(string)
	log.Printf("Reading web deployment configuration %s", d.Id())
	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		if version == "" {
			version = wp.determineLatestVersion(ctx, d.Id())
		}
		configuration, resp, getErr := wp.getWebdeploymentsConfigurationVersion(ctx, d.Id(), version)

		if getErr != nil {
			if gcloud.IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read web deployment configuration %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read web deployment configuration %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceWebDeploymentConfiguration())
		d.Set("name", *configuration.Name)

		if configuration.Description != nil {
			d.Set("description", *configuration.Description)
		}
		if configuration.Languages != nil {
			d.Set("languages", *configuration.Languages)
		}
		if configuration.DefaultLanguage != nil {
			d.Set("default_language", *configuration.DefaultLanguage)
		}
		if configuration.Status != nil {
			d.Set("status", *configuration.Status)
		}
		if configuration.Version != nil {
			d.Set("version", *configuration.Version)
		}
		if configuration.Messenger != nil {
			d.Set("messenger", flattenMessengerSettings(configuration.Messenger))
		}
		if configuration.Cobrowse != nil {
			d.Set("cobrowse", flattenCobrowseSettings(configuration.Cobrowse))
		}
		if configuration.JourneyEvents != nil {
			d.Set("journey_events", flattenJourneyEvents(configuration.JourneyEvents))
		}

		log.Printf("Read web deployment configuration %s %s", d.Id(), *configuration.Name)
		return cc.CheckState()
	})
}

func updateWebDeploymentConfiguration(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	wp := getWebDeploymentConfigurationsProxy(sdkConfig)
	name, inputCfg := readWebDeploymentConfigurationFromResourceData(d)

	log.Printf("Updating web deployment configuration %s", name)

	diagErr := gcloud.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := wp.updateWebdeploymentsConfigurationVersionsDraft(ctx, d.Id(), *inputCfg)
		if err != nil {
			if gcloud.IsStatus400(resp) {
				return retry.RetryableError(fmt.Errorf("Error updating web deployment configuration %s: %s", name, err))
			}
			return retry.NonRetryableError(fmt.Errorf("Error updating web deployment configuration %s: %s", name, err))
		}
		return nil
	})
	if diagErr != nil {
		return diagErr
	}

	activeError := waitForConfigurationDraftToBeActive(ctx, meta, d.Id())
	if activeError != nil {
		return diag.Errorf("Web deployment configuration %s did not become active and could not be published", name)
	}

	diagErr = gcloud.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		configuration, resp, err := wp.createWebdeploymentsConfigurationVersionsDraftPublish(ctx, d.Id())
		if err != nil {
			if gcloud.IsStatus400(resp) {
				return retry.RetryableError(fmt.Errorf("Error publishing web deployment configuration %s: %s", name, err))
			}
			return retry.NonRetryableError(fmt.Errorf("Error publishing web deployment configuration %s: %s", name, err))
		}
		d.Set("version", configuration.Version)
		d.Set("status", configuration.Status)
		return nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Finished updating web deployment configuration %s", name)
	return readWebDeploymentConfiguration(ctx, d, meta)
}

func deleteWebDeploymentConfiguration(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	wp := getWebDeploymentConfigurationsProxy(sdkConfig)

	log.Printf("Deleting web deployment configuration %s", name)
	_, err := wp.deleteWebDeploymentConfiguration(ctx, d.Id())

	if err != nil {
		return diag.Errorf("Failed to delete web deployment configuration %s: %s", name, err)
	}

	return gcloud.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {

		_, resp, err := wp.getWebdeploymentsConfigurationVersionsDraft(ctx, d.Id())
		if err != nil {
			if gcloud.IsStatus404(resp) {
				log.Printf("Deleted web deployment configuration %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting web deployment configuration %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("Web deployment configuration %s still exists", d.Id()))
	})
}
