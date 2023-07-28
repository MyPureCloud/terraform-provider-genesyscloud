package webdeployments_configuration

import (
	"context"
	"fmt"
	"log"
	"time"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

func getAllWebDeploymentConfigurations(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	webDeploymentsAPI := platformclientv2.NewWebDeploymentsApiWithConfig(clientConfig)

	configurations, _, getErr := webDeploymentsAPI.GetWebdeploymentsConfigurations(false)
	if getErr != nil {
		return nil, diag.Errorf("Failed to get web deployment configurations: %v", getErr)
	}

	for _, configuration := range *configurations.Entities {
		resources[*configuration.Id] = &resourceExporter.ResourceMeta{Name: *configuration.Name}
	}

	return resources, nil
}

func createWebDeploymentConfiguration(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name, inputCfg := readWebDeploymentConfigurationFromResourceData(d)

	log.Printf("Creating web deployment configuration %s", name)

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	api := platformclientv2.NewWebDeploymentsApiWithConfig(sdkConfig)

	diagErr := gcloud.WithRetries(ctx, 30*time.Second, func() *resource.RetryError {
		configuration, resp, err := api.PostWebdeploymentsConfigurations(*inputCfg)
		if err != nil {
			if gcloud.IsStatus400(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to create web deployment configuration %s: %s", name, err))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to create web deployment configuration %s: %s", name, err))
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

	diagErr = gcloud.WithRetries(ctx, 30*time.Second, func() *resource.RetryError {
		configuration, resp, err := api.PostWebdeploymentsConfigurationVersionsDraftPublish(d.Id())
		if err != nil {
			if gcloud.IsStatus400(resp) {
				return resource.RetryableError(fmt.Errorf("Error publishing web deployment configuration %s: %s", name, err))
			}
			return resource.NonRetryableError(fmt.Errorf("Error publishing web deployment configuration %s: %s", name, err))
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
	api := platformclientv2.NewWebDeploymentsApiWithConfig(sdkConfig)

	version := d.Get("version").(string)
	log.Printf("Reading web deployment configuration %s", d.Id())
	return gcloud.WithRetriesForRead(ctx, d, func() *resource.RetryError {
		if version == "" {
			version = determineLatestVersion(ctx, api, d.Id())
		}
		configuration, resp, getErr := api.GetWebdeploymentsConfigurationVersion(d.Id(), version)
		if getErr != nil {
			if gcloud.IsStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to read web deployment configuration %s: %s", d.Id(), getErr))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to read web deployment configuration %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceWebDeploymentConfiguration())

		d.Set("name", *configuration.Name)
		d.Set("languages", *configuration.Languages)
		d.Set("default_language", *configuration.DefaultLanguage)

		resourcedata.SetNillableValue(d, "description", configuration.Description)
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

	diagErr := gcloud.WithRetries(ctx, 30*time.Second, func() *resource.RetryError {
		_, resp, err := api.PutWebdeploymentsConfigurationVersionsDraft(d.Id(), *inputCfg)
		if err != nil {
			if gcloud.IsStatus400(resp) {
				return resource.RetryableError(fmt.Errorf("Error updating web deployment configuration %s: %s", name, err))
			}
			return resource.NonRetryableError(fmt.Errorf("Error updating web deployment configuration %s: %s", name, err))
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

	diagErr = gcloud.WithRetries(ctx, 30*time.Second, func() *resource.RetryError {
		configuration, resp, err := api.PostWebdeploymentsConfigurationVersionsDraftPublish(d.Id())
		if err != nil {
			if gcloud.IsStatus400(resp) {
				return resource.RetryableError(fmt.Errorf("Error publishing web deployment configuration %s: %s", name, err))
			}
			return resource.NonRetryableError(fmt.Errorf("Error publishing web deployment configuration %s: %s", name, err))
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

	log.Printf("Deleting web deployment configuration %s", name)
	_, err := api.DeleteWebdeploymentsConfiguration(d.Id())

	if err != nil {
		return diag.Errorf("Failed to delete web deployment configuration %s: %s", name, err)
	}

	return gcloud.WithRetries(ctx, 30*time.Second, func() *resource.RetryError {
		_, resp, err := api.GetWebdeploymentsConfigurationVersionsDraft(d.Id())
		if err != nil {
			if gcloud.IsStatus404(resp) {
				log.Printf("Deleted web deployment configuration %s", d.Id())
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("Error deleting web deployment configuration %s: %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("Web deployment configuration %s still exists", d.Id()))
	})
}
