package webdeployments_configuration

import (
	"context"
	"fmt"
	"log"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	wdcUtils "terraform-provider-genesyscloud/genesyscloud/webdeployments_configuration/utils"

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
				return retry.RetryableError(fmt.Errorf("error verifying active status for new web deployment configuration %s: %s", id, err))
			}
			return retry.NonRetryableError(fmt.Errorf("error verifying active status for new web deployment configuration %s: %s", id, err))
		}

		if *configuration.Status == "Active" {
			return nil
		}

		return retry.RetryableError(fmt.Errorf("web deployment configuration %s not active yet. Status: %s", id, *configuration.Status))
	})
}

func createWebDeploymentConfiguration(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	wp := getWebDeploymentConfigurationsProxy(sdkConfig)

	name, inputCfg := wdcUtils.ReadWebDeploymentConfigurationFromResourceData(d)
	log.Printf("Creating web deployment configuration %s", name)

	diagErr := gcloud.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		configuration, resp, err := wp.createWebdeploymentsConfiguration(ctx, *inputCfg)
		if err != nil {
			var extraErrorInfo string
			featureIsNotImplemented, fieldName := wdcUtils.FeatureNotImplemented(resp)
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
				return retry.RetryableError(fmt.Errorf("error publishing web deployment configuration %s: %s", name, err))
			}
			return retry.NonRetryableError(fmt.Errorf("error publishing web deployment configuration %s: %s", name, err))
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
				return retry.RetryableError(fmt.Errorf("failed to read web deployment configuration %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("failed to read web deployment configuration %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceWebDeploymentConfiguration())
		d.Set("name", *configuration.Name)

		resourcedata.SetNillableValue(d, "description", configuration.Description)
		resourcedata.SetNillableValue(d, "languages", configuration.Languages)
		resourcedata.SetNillableValue(d, "default_language", configuration.DefaultLanguage)
		resourcedata.SetNillableValue(d, "status", configuration.Status)
		resourcedata.SetNillableValue(d, "version", configuration.Version)
		if configuration.HeadlessMode != nil {
			resourcedata.SetNillableValue(d, "headless_mode_enabled", configuration.HeadlessMode.Enabled)
		}

		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "custom_i18n_labels", configuration.CustomI18nLabels, wdcUtils.FlattenCustomI18nLabels)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "position", configuration.Position, wdcUtils.FlattenPosition)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "authentication_settings", configuration.AuthenticationSettings, wdcUtils.FlattenAuthenticationSettings)

		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "messenger", configuration.Messenger, wdcUtils.FlattenMessengerSettings)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "cobrowse", configuration.Cobrowse, wdcUtils.FlattenCobrowseSettings)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "journey_events", configuration.JourneyEvents, wdcUtils.FlattenJourneyEvents)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "support_center", configuration.SupportCenter, wdcUtils.FlattenSupportCenterSettings)

		log.Printf("Read web deployment configuration %s %s", d.Id(), *configuration.Name)
		return cc.CheckState()
	})
}

func updateWebDeploymentConfiguration(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	wp := getWebDeploymentConfigurationsProxy(sdkConfig)
	name, inputCfg := wdcUtils.ReadWebDeploymentConfigurationFromResourceData(d)

	log.Printf("Updating web deployment configuration %s", name)

	diagErr := gcloud.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := wp.updateWebdeploymentsConfigurationVersionsDraft(ctx, d.Id(), *inputCfg)
		if err != nil {
			if gcloud.IsStatus400(resp) {
				return retry.RetryableError(fmt.Errorf("error updating web deployment configuration %s: %s", name, err))
			}
			return retry.NonRetryableError(fmt.Errorf("error updating web deployment configuration %s: %s", name, err))
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
				return retry.RetryableError(fmt.Errorf("error publishing web deployment configuration %s: %s", name, err))
			}
			return retry.NonRetryableError(fmt.Errorf("error publishing web deployment configuration %s: %s", name, err))
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
			return retry.NonRetryableError(fmt.Errorf("error deleting web deployment configuration %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("web deployment configuration %s still exists", d.Id()))
	})
}
