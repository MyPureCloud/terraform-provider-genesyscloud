package webdeployments_configuration

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
	wdcUtils "terraform-provider-genesyscloud/genesyscloud/webdeployments_configuration/utils"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

func getAllWebDeploymentConfigurations(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	wp := getWebDeploymentConfigurationsProxy(clientConfig)

	configurations, resp, err := wp.getWebDeploymentsConfiguration(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get webdeployments configuration error: %s", err), resp)
	}
	for _, configuration := range *configurations.Entities {
		resources[*configuration.Id] = &resourceExporter.ResourceMeta{BlockLabel: *configuration.Name}
	}
	return resources, nil
}

func waitForConfigurationDraftToBeActive(ctx context.Context, meta interface{}, id string) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	wp := getWebDeploymentConfigurationsProxy(sdkConfig)

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		configuration, resp, err := wp.getWebdeploymentsConfigurationVersionsDraft(ctx, id)
		if err != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error verifying active status for new web deployment configuration %s | error: %s", id, err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error verifying active status for new web deployment configuration %s | error: %s", id, err), resp))
		}

		if *configuration.Status == "Active" {
			return nil
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("web deployment configuration %s not active yet. Status: %s", id, *configuration.Status), resp))
	})
}

func createWebDeploymentConfiguration(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	wp := getWebDeploymentConfigurationsProxy(sdkConfig)

	name, inputCfg := wdcUtils.BuildWebDeploymentConfigurationFromResourceData(d)
	log.Printf("Creating web deployment configuration %s", name)
	log.Println("Current Deployment: ", inputCfg)
	diagErr := util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		configuration, resp, err := wp.createWebdeploymentsConfiguration(ctx, *inputCfg)
		if err != nil {
			var extraErrorInfo string
			featureIsNotImplemented, fieldName := wdcUtils.FeatureNotImplemented(resp)
			if featureIsNotImplemented {
				extraErrorInfo = fmt.Sprintf("Feature '%s' is not yet implemented", fieldName)
			}
			if util.IsStatus400(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to create web deployment configuration %s: %s. %s", name, err, extraErrorInfo), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to create web deployment configuration %s: %s. %s", name, err, extraErrorInfo), resp))
		}
		d.SetId(*configuration.Id)
		_ = d.Set("status", configuration.Status)

		return nil
	})
	if diagErr != nil {
		return diagErr
	}

	activeError := waitForConfigurationDraftToBeActive(ctx, meta, d.Id())
	if activeError != nil {
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Web deployment configuration %s did not become active and could not be published", name), fmt.Errorf("%v", activeError))
	}

	diagErr = util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		configuration, resp, err := wp.createWebdeploymentsConfigurationVersionsDraftPublish(ctx, d.Id())
		if err != nil {
			if util.IsStatus400(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error publishing web deployment configuration %s | error: %s", name, err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error publishing web deployment configuration %s | error: %s", name, err), resp))
		}
		_ = d.Set("version", configuration.Version)
		_ = d.Set("status", configuration.Status)
		log.Printf("Created web deployment configuration %s %s", name, *configuration.Id)

		return nil
	})
	if diagErr != nil {
		return diagErr
	}

	return readWebDeploymentConfiguration(ctx, d, meta)
}

func readWebDeploymentConfiguration(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	wp := getWebDeploymentConfigurationsProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceWebDeploymentConfiguration(), constants.ConsistencyChecks(), ResourceType)

	version := d.Get("version").(string)
	log.Printf("Reading web deployment configuration %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		if version == "" {
			version = wp.determineLatestVersion(ctx, d.Id())
		}
		configuration, resp, getErr := wp.getWebdeploymentsConfigurationVersion(ctx, d.Id(), version)

		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read web deployment configuration %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read web deployment configuration %s | error: %s", d.Id(), getErr), resp))
		}

		_ = d.Set("name", *configuration.Name)

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
		return cc.CheckState(d)
	})
}

func updateWebDeploymentConfiguration(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	wp := getWebDeploymentConfigurationsProxy(sdkConfig)
	name, inputCfg := wdcUtils.BuildWebDeploymentConfigurationFromResourceData(d)

	log.Printf("Updating web deployment configuration %s", name)

	diagErr := util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := wp.updateWebdeploymentsConfigurationVersionsDraft(ctx, d.Id(), *inputCfg)
		if err != nil {
			if util.IsStatus400(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error updating web deployment configuration %s | error: %s", name, err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error updating web deployment configuration %s | error: %s", name, err), resp))
		}
		return nil
	})
	if diagErr != nil {
		return diagErr
	}

	activeError := waitForConfigurationDraftToBeActive(ctx, meta, d.Id())
	if activeError != nil {
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Web deployment configuration %s did not become active and could not be published", name), fmt.Errorf("%v", activeError))
	}

	diagErr = util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		configuration, resp, err := wp.createWebdeploymentsConfigurationVersionsDraftPublish(ctx, d.Id())
		if err != nil {
			if util.IsStatus400(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error publishing web deployment configuration %s | error: %s", name, err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error publishing web deployment configuration %s | error: %s", name, err), resp))
		}
		_ = d.Set("version", configuration.Version)
		_ = d.Set("status", configuration.Status)
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

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	wp := getWebDeploymentConfigurationsProxy(sdkConfig)

	log.Printf("Deleting web deployment configuration %s", name)
	resp, err := wp.deleteWebDeploymentConfiguration(ctx, d.Id())

	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete web deployment configuration %s error: %s", name, err), resp)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {

		_, resp, err := wp.getWebdeploymentsConfigurationVersionsDraft(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted web deployment configuration %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error deleting web deployment configuration %s | error: %s", d.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Web deployment configuration %s still exists", d.Id()), resp))
	})
}
