package webdeployments_configuration

import (
	"context"
	"fmt"
	"log"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
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

func waitForConfigurationDraftToBeActive(ctx context.Context, api *platformclientv2.WebDeploymentsApi, id string) diag.Diagnostics {
	return gcloud.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		configuration, resp, err := api.GetWebdeploymentsConfigurationVersionsDraft(id)
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

func readWebDeploymentConfigurationFromResourceData(d *schema.ResourceData) (string, *platformclientv2.Webdeploymentconfigurationversion) {
	name := d.Get("name").(string)
	languages := lists.InterfaceListToStrings(d.Get("languages").([]interface{}))
	defaultLanguage := d.Get("default_language").(string)

	inputCfg := &platformclientv2.Webdeploymentconfigurationversion{
		Name:            &name,
		Languages:       &languages,
		DefaultLanguage: &defaultLanguage,
	}

	description, ok := d.Get("description").(string)
	if ok {
		inputCfg.Description = &description
	}

	messengerSettings := readMessengerSettings(d)
	if messengerSettings != nil {
		inputCfg.Messenger = messengerSettings
	}

	cobrowseSettings := readCobrowseSettings(d)
	if cobrowseSettings != nil {
		inputCfg.Cobrowse = cobrowseSettings
	}

	journeySettings := readJourneySettings(d)
	if journeySettings != nil {
		inputCfg.JourneyEvents = journeySettings
	}

	return name, inputCfg
}

func readJourneySettings(d *schema.ResourceData) *platformclientv2.Journeyeventssettings {
	value, ok := d.GetOk("journey_events")
	if !ok {
		return nil
	}

	cfgs := value.([]interface{})
	if len(cfgs) < 1 {
		return nil
	}

	cfg := cfgs[0].(map[string]interface{})
	enabled, _ := cfg["enabled"].(bool)
	journeySettings := &platformclientv2.Journeyeventssettings{
		Enabled: &enabled,
	}

	excludedQueryParams := lists.InterfaceListToStrings(cfg["excluded_query_parameters"].([]interface{}))
	journeySettings.ExcludedQueryParameters = &excludedQueryParams

	if keepUrlFragment, ok := cfg["should_keep_url_fragment"].(bool); ok && keepUrlFragment {
		journeySettings.ShouldKeepUrlFragment = &keepUrlFragment
	}

	searchQueryParameters := lists.InterfaceListToStrings(cfg["search_query_parameters"].([]interface{}))
	journeySettings.SearchQueryParameters = &searchQueryParameters

	pageviewConfig := cfg["pageview_config"]
	if value, ok := pageviewConfig.(string); ok {
		if value != "" {
			journeySettings.PageviewConfig = &value
		}
	}

	if clickEvents := readSelectorEventTriggers(cfg["click_event"].([]interface{})); clickEvents != nil {
		journeySettings.ClickEvents = clickEvents
	}

	if formsTrackEvents := readFormsTrackTriggers(cfg["form_track_event"].([]interface{})); formsTrackEvents != nil {
		journeySettings.FormsTrackEvents = formsTrackEvents
	}

	if idleEvents := readIdleEventTriggers(cfg["idle_event"].([]interface{})); idleEvents != nil {
		journeySettings.IdleEvents = idleEvents
	}

	if inViewportEvents := readSelectorEventTriggers(cfg["in_viewport_event"].([]interface{})); inViewportEvents != nil {
		journeySettings.InViewportEvents = inViewportEvents
	}

	if scrollDepthEvents := readScrollPercentageEventTriggers(cfg["scroll_depth_event"].([]interface{})); scrollDepthEvents != nil {
		journeySettings.ScrollDepthEvents = scrollDepthEvents
	}

	return journeySettings
}

func createWebDeploymentConfiguration(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name, inputCfg := readWebDeploymentConfigurationFromResourceData(d)

	log.Printf("Creating web deployment configuration %s", name)

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	api := platformclientv2.NewWebDeploymentsApiWithConfig(sdkConfig)

	diagErr := gcloud.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		configuration, resp, err := api.PostWebdeploymentsConfigurations(*inputCfg)
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

	activeError := waitForConfigurationDraftToBeActive(ctx, api, d.Id())
	if activeError != nil {
		return diag.Errorf("Web deployment configuration %s did not become active and could not be published", name)
	}

	diagErr = gcloud.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		configuration, resp, err := api.PostWebdeploymentsConfigurationVersionsDraftPublish(d.Id())
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

		resourcedata.SetNillableValue(d, "description", configuration.Description)
		resourcedata.SetNillableValue(d, "languages", configuration.Languages)
		resourcedata.SetNillableValue(d, "default_language", configuration.DefaultLanguage)
		resourcedata.SetNillableValue(d, "status", configuration.Status)
		resourcedata.SetNillableValue(d, "version", configuration.Version)

		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "messenger", configuration.Messenger, flattenMessengerSettings)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "cobrowse", configuration.Cobrowse, flattenCobrowseSettings)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "journey_events", configuration.JourneyEvents, flattenJourneyEvents)

		log.Printf("Read web deployment configuration %s %s", d.Id(), *configuration.Name)
		return cc.CheckState()
	})
}

func updateWebDeploymentConfiguration(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name, inputCfg := readWebDeploymentConfigurationFromResourceData(d)

	log.Printf("Updating web deployment configuration %s", name)

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	api := platformclientv2.NewWebDeploymentsApiWithConfig(sdkConfig)

	diagErr := gcloud.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := api.PutWebdeploymentsConfigurationVersionsDraft(d.Id(), *inputCfg)
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

	activeError := waitForConfigurationDraftToBeActive(ctx, api, d.Id())
	if activeError != nil {
		return diag.Errorf("Web deployment configuration %s did not become active and could not be published", name)
	}

	diagErr = gcloud.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		configuration, resp, err := api.PostWebdeploymentsConfigurationVersionsDraftPublish(d.Id())
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
	api := platformclientv2.NewWebDeploymentsApiWithConfig(sdkConfig)
	wp := getWebDeploymentConfigurationsProxy(sdkConfig)

	log.Printf("Deleting web deployment configuration %s", name)
	_, err := wp.deleteWebDeploymentConfiguration(ctx, d.Id())

	if err != nil {
		return diag.Errorf("Failed to delete web deployment configuration %s: %s", name, err)
	}

	return gcloud.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {

		//TODO STOPPED HERE
		_, resp, err := api.GetWebdeploymentsConfigurationVersionsDraft(d.Id())
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
