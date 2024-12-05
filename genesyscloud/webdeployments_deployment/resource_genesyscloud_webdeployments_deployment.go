package webdeployments_deployment

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

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

func getAllWebDeployments(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	wd := getWebDeploymentsProxy(clientConfig)

	deployments, resp, getErr := wd.getWebDeployments(ctx)
	if getErr != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get web deployments error: %s", getErr), resp)
	}

	for _, deployment := range *deployments.Entities {
		resources[*deployment.Id] = &resourceExporter.ResourceMeta{BlockLabel: *deployment.Name}
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
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Failed to create web deployment %s", name), err)
	}

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	wd := getWebDeploymentsProxy(sdkConfig)

	log.Printf("Creating web deployment %s", name)

	configId := d.Get("configuration.0.id").(string)
	inputConfigVersion := d.Get("configuration.0.version").(string)

	flow := util.BuildSdkWebdeploymentFlowEntityRef(d, "flow_id")

	configVersion, versionList, er := wd.determineLatestVersion(ctx, configId)
	if er != nil {
		return er
	}
	if inputConfigVersion == "" {
		inputConfigVersion = configVersion
	}

	exists := util.StringExists(inputConfigVersion, versionList)
	if !exists {
		log.Printf("For Web deployment Resource %v, Configuration Version Input %v does not match with any existing versions %v",
			name, inputConfigVersion, versionList)
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("For Web Deployment Resource %v, Configuration Version Input %v does not match with any existing versions %v", name, inputConfigVersion, versionList), nil)
	}

	inputDeployment := platformclientv2.Webdeployment{
		Name: &name,
		Configuration: &platformclientv2.Webdeploymentconfigurationversionentityref{
			Id:      &configId,
			Version: &inputConfigVersion,
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

	diagErr := util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		deployment, resp, err := wd.createWebDeployment(ctx, inputDeployment)
		if err != nil {
			if util.IsStatus400(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to create web deployment %s | error: %s", name, err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to create web deployment %s | error: %s", name, err), resp))
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
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Web deployment %s did not become active and could not be created", name), fmt.Errorf("%v", activeError))
	}
	return readWebDeployment(ctx, d, meta)
}

func waitForDeploymentToBeActive(ctx context.Context, sdkConfig *platformclientv2.Configuration, id string) diag.Diagnostics {
	wd := getWebDeploymentsProxy(sdkConfig)
	return util.WithRetries(ctx, 60*time.Second, func() *retry.RetryError {
		deployment, resp, err := wd.getWebDeployment(ctx, id)
		if err != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error verifying active status for new web deployment %s | error: %s", id, err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error verifying active status for new web deployment %s | error: %s", id, err), resp))
		}

		if *deployment.Status == "Active" {
			return nil
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Web deployment %s not active yet | Status: %s", id, *deployment.Status), resp))
	})
}

func readWebDeployment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	wd := getWebDeploymentsProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceWebDeployment(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading web deployment %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		deployment, resp, getErr := wd.getWebDeployment(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read web deployment %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read web deployment %s | error: %s", d.Id(), getErr), resp))
		}

		_ = d.Set("name", *deployment.Name)

		if deployment.Description != nil {
			_ = d.Set("description", *deployment.Description)
		}
		if deployment.AllowAllDomains != nil {
			_ = d.Set("allow_all_domains", *deployment.AllowAllDomains)
		}
		_ = d.Set("configuration", flattenConfiguration(deployment.Configuration))
		if deployment.AllowedDomains != nil && len(*deployment.AllowedDomains) > 0 {
			_ = d.Set("allowed_domains", *deployment.AllowedDomains)
		}
		if deployment.Flow != nil {
			_ = d.Set("flow_id", *deployment.Flow.Id)
		}
		if deployment.Status != nil {
			_ = d.Set("status", *deployment.Status)
		}

		log.Printf("Read web deployment %s %s", d.Id(), *deployment.Name)
		return cc.CheckState(d)
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
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Failed to update web deployment %s", name), err)
	}

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	wd := getWebDeploymentsProxy(sdkConfig)

	log.Printf("Updating web deployment %s", name)

	configId := d.Get("configuration.0.id").(string)

	flow := util.BuildSdkWebdeploymentFlowEntityRef(d, "flow_id")

	// always update to latest version of configuration during update of an existing webdeployment
	configVersion, _, er := wd.determineLatestVersion(ctx, configId)
	if err != nil {
		return er
	}
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

	diagErr := util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := wd.updateWebDeployment(ctx, d.Id(), inputDeployment)
		if err != nil {
			if util.IsStatus400(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error updating web deployment %s | error: %s", name, err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error updating web deployment %s | error: %s", name, err), resp))
		}

		return nil
	})
	if diagErr != nil {
		return diagErr
	}

	activeError := waitForDeploymentToBeActive(ctx, sdkConfig, d.Id())

	if activeError != nil {
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Web deployment %s did not become active and could not be created", name), fmt.Errorf("%v", activeError))
	}

	log.Printf("Finished updating web deployment %s", name)
	return readWebDeployment(ctx, d, meta)
}

func deleteWebDeployment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	wd := getWebDeploymentsProxy(sdkConfig)

	log.Printf("Deleting web deployment %s", name)
	resp, err := wd.deleteWebDeployment(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete web deployment %s error: %s", name, err), resp)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := wd.getWebDeployment(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted web deployment %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting web deployment %s | error: %s", d.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Web deployment %s still exists", d.Id()), resp))
	})
}
