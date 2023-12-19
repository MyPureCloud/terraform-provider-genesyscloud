package webdeployments_deployment

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"
)

func getAllWebDeployments(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	wd := getWebDeploymentsProxy(clientConfig)

	deployments, _, getErr := wd.getWebDeployments(ctx)
	if getErr != nil {
		return nil, diag.Errorf("Failed to get web deployments: %v", getErr)
	}

	for _, deployment := range *deployments.Entities {
		resources[*deployment.Id] = &resourceExporter.ResourceMeta{Name: *deployment.Name}
	}

	return resources, nil
}

func createWebDeployment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	allowAllDomains := d.Get("allow_all_domains").(bool)
	allowedDomains := lists.InterfaceListToStrings(d.Get("allowed_domains").([]interface{}))

	err := validAllowedDomainsSettings(d)
	if err != nil {
		return diag.Errorf("Failed to create web deployment %s: %s", name, err)
	}

	sdkConfig := meta.(*genesyscloud.ProviderMeta).ClientConfig
	wd := getWebDeploymentsProxy(sdkConfig)

	log.Printf("Creating web deployment %s", name)

	configId := d.Get("configuration.0.id").(string)
	configVersion := d.Get("configuration.0.version").(string)

	flow := genesyscloud.BuildSdkDomainEntityRef(d, "flow_id")

	inputDeployment := platformclientv2.Webdeployment{
		Name: &name,
		Configuration: &platformclientv2.Webdeploymentconfigurationversionentityref{
			Id:      &configId,
			Version: &configVersion,
		},
		AllowAllDomains: &allowAllDomains,
		AllowedDomains:  &allowedDomains,
	}

	if description != "" {
		inputDeployment.Description = &description
	}

	if flow != nil {
		inputDeployment.Flow = flow
	}

	diagErr := genesyscloud.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		deployment, resp, err := wd.createWebDeployment(ctx, inputDeployment)
		if err != nil {
			if genesyscloud.IsStatus400(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to create web deployment %s: %s", name, err))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to create web deployment %s: %s", name, err))
		}

		d.SetId(*deployment.Id)

		log.Printf("Created web deployment %s %s %s", name, *deployment.Id, resp.CorrelationID)

		return nil
	})
	if diagErr != nil {
		return diagErr
	}

	time.Sleep(10 * time.Second)
	activeError := waitForDeploymentToBeActive(ctx, sdkConfig, d.Id())
	if activeError != nil {
		return diag.Errorf("Web deployment %s did not become active and could not be created", name)
	}

	return readWebDeployment(ctx, d, meta)
}

func waitForDeploymentToBeActive(ctx context.Context, sdkConfig *platformclientv2.Configuration, id string) diag.Diagnostics {
	wd := getWebDeploymentsProxy(sdkConfig)
	return genesyscloud.WithRetries(ctx, 60*time.Second, func() *retry.RetryError {
		deployment, resp, err := wd.getWebDeployment(ctx, id)
		if err != nil {
			if genesyscloud.IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Error verifying active status for new web deployment %s: %s", id, err))
			}
			return retry.NonRetryableError(fmt.Errorf("Error verifying active status for new web deployment %s: %s", id, err))
		}

		if *deployment.Status == "Active" {
			return nil
		}

		return retry.RetryableError(fmt.Errorf("Web deployment %s not active yet. Status: %s", id, *deployment.Status))
	})
}

func readWebDeployment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*genesyscloud.ProviderMeta).ClientConfig
	wd := getWebDeploymentsProxy(sdkConfig)

	log.Printf("Reading web deployment %s", d.Id())
	return genesyscloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		deployment, resp, getErr := wd.getWebDeployment(ctx, d.Id())
		if getErr != nil {
			if genesyscloud.IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read web deployment %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read web deployment %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceWebDeployment())
		d.Set("name", *deployment.Name)

		if deployment.Description != nil {
			d.Set("description", *deployment.Description)
		}
		if deployment.AllowAllDomains != nil {
			d.Set("allow_all_domains", *deployment.AllowAllDomains)
		}
		d.Set("configuration", flattenConfiguration(deployment.Configuration))
		if deployment.AllowedDomains != nil && len(*deployment.AllowedDomains) > 0 {
			d.Set("allowed_domains", *deployment.AllowedDomains)
		}
		if deployment.Flow != nil {
			d.Set("flow_id", *deployment.Flow.Id)
		}
		if deployment.Status != nil {
			d.Set("status", *deployment.Status)
		}

		log.Printf("Read web deployment %s %s", d.Id(), *deployment.Name)
		return cc.CheckState()
	})
}

func updateWebDeployment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	allowAllDomains := d.Get("allow_all_domains").(bool)
	allowedDomains := lists.InterfaceListToStrings(d.Get("allowed_domains").([]interface{}))
	status := d.Get("status").(string)

	err := validAllowedDomainsSettings(d)
	if err != nil {
		return diag.Errorf("Failed to update web deployment %s: %s", name, err)
	}

	sdkConfig := meta.(*genesyscloud.ProviderMeta).ClientConfig
	wd := getWebDeploymentsProxy(sdkConfig)

	log.Printf("Updating web deployment %s", name)

	configId := d.Get("configuration.0.id").(string)
	configVersion := d.Get("configuration.0.version").(string)

	flow := genesyscloud.BuildSdkDomainEntityRef(d, "flow_id")

	inputDeployment := platformclientv2.Webdeployment{
		Name: &name,
		Configuration: &platformclientv2.Webdeploymentconfigurationversionentityref{
			Id:      &configId,
			Version: &configVersion,
		},
		AllowAllDomains: &allowAllDomains,
		AllowedDomains:  &allowedDomains,
		Status:          &status,
	}

	if description != "" {
		inputDeployment.Description = &description
	}

	if flow != nil {
		inputDeployment.Flow = flow
	}

	diagErr := genesyscloud.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := wd.updateWebDeployment(ctx, d.Id(), inputDeployment)
		if err != nil {
			if genesyscloud.IsStatus400(resp) {
				return retry.RetryableError(fmt.Errorf("Error updating web deployment %s: %s", name, err))
			}
			return retry.NonRetryableError(fmt.Errorf("Error updating web deployment %s: %s", name, err))
		}

		return nil
	})
	if diagErr != nil {
		return diagErr
	}

	activeError := waitForDeploymentToBeActive(ctx, sdkConfig, d.Id())
	if activeError != nil {
		return diag.Errorf("Web deployment %s did not become active and could not be created", name)
	}

	log.Printf("Finished updating web deployment %s", name)
	return readWebDeployment(ctx, d, meta)
}

func deleteWebDeployment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*genesyscloud.ProviderMeta).ClientConfig
	wd := getWebDeploymentsProxy(sdkConfig)

	log.Printf("Deleting web deployment %s", name)
	_, err := wd.deleteWebDeployment(ctx, d.Id())

	if err != nil {
		return diag.Errorf("Failed to delete web deployment %s: %s", name, err)
	}

	return genesyscloud.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := wd.getWebDeployment(ctx, d.Id())
		if err != nil {
			if genesyscloud.IsStatus404(resp) {
				log.Printf("Deleted web deployment %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting web deployment %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("Web deployment %s still exists", d.Id()))
	})
}
