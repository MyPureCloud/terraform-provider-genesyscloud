package webdeployments_deployment_identity_resolution

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
	consistencyChecker "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
)

func getAllWebDeploymentIdentityResolutions(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getWebDeploymentIdentityResolutionProxy(clientConfig)

	deployments, resp, err := proxy.webDeploymentsProxy.GetWebDeployments(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, "failed to list web deployments for identity resolution export", resp)
	}
	if deployments == nil || deployments.Entities == nil {
		return resources, nil
	}

	for _, deployment := range *deployments.Entities {
		if deployment.Id == nil || deployment.Name == nil {
			continue
		}

		config, getResp, getErr := proxy.getWebDeploymentIdentityResolution(ctx, *deployment.Id)
		if getErr != nil {
			if util.IsStatus404(getResp) {
				continue
			}
			return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to get identity resolution for deployment %s", *deployment.Id), getResp)
		}

		if isDefaultIdentityResolutionConfig(config) {
			continue
		}

		resources[*deployment.Id] = &resourceExporter.ResourceMeta{BlockLabel: *deployment.Name + "-identity-resolution"}
	}

	return resources, nil
}

func createWebDeploymentIdentityResolution(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	deploymentId := d.Get("deployment_id").(string)

	log.Printf("creating identity resolution for deployment %s", deploymentId)
	d.SetId(deploymentId)

	return updateWebDeploymentIdentityResolution(ctx, d, meta)
}

func readWebDeploymentIdentityResolution(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getWebDeploymentIdentityResolutionProxy(sdkConfig)
	cc := consistencyChecker.NewConsistencyCheck(ctx, d, meta, ResourceWebDeploymentIdentityResolution(), constants.ConsistencyChecks(), ResourceType)

	deploymentId := d.Id()
	log.Printf("reading identity resolution for deployment %s", deploymentId)

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		config, resp, getErr := proxy.getWebDeploymentIdentityResolution(ctx, deploymentId)
		if getErr != nil {
			if util.IsStatus404(resp) {
				log.Printf("parent deployment %s not found, removing identity resolution from state", deploymentId)
				d.SetId("")
				return nil
			}
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read identity resolution for deployment %s: %s", deploymentId, getErr), resp))
		}

		_ = d.Set("deployment_id", deploymentId)
		if config.ResolveIdentities != nil {
			_ = d.Set("resolve_identities", *config.ResolveIdentities)
		} else {
			_ = d.Set("resolve_identities", true)
		}
		if config.Division != nil && config.Division.Id != nil && !isUnassignedDivisionId(*config.Division.Id) {
			_ = d.Set("division_id", *config.Division.Id)
		} else {
			_ = d.Set("division_id", "")
		}
		if config.ExternalSource != nil && config.ExternalSource.Id != nil {
			_ = d.Set("external_source_id", *config.ExternalSource.Id)
		} else {
			_ = d.Set("external_source_id", "")
		}

		automergeDeclared := len(d.Get("automerge_config").([]interface{})) > 0
		_ = d.Set("automerge_config", flattenAutomergeConfig(config.Automerge, automergeDeclared))

		log.Printf("read identity resolution for deployment %s", deploymentId)
		return cc.CheckState(d)
	})
}

func updateWebDeploymentIdentityResolution(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getWebDeploymentIdentityResolutionProxy(sdkConfig)
	deploymentId := d.Id()

	config, err := buildDeploymentIdentityResolutionConfig(d)
	if err != nil {
		return util.BuildDiagnosticError(ResourceType, "failed to build identity resolution config", err)
	}

	log.Printf("updating identity resolution for deployment %s", deploymentId)
	stored, resp, putErr := proxy.putWebDeploymentIdentityResolution(ctx, deploymentId, config)
	if putErr != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to update identity resolution for deployment %s", deploymentId), resp)
	}

	// The PUT answers 200 even when the platform refuses to enable authenticated web
	// messaging automerge, which would otherwise show up only as a plan that never
	// converges. Fail loudly instead.
	if authenticatedWebMessagingDowngraded(&config, stored) {
		return util.BuildDiagnosticError(ResourceType,
			fmt.Sprintf("the platform did not enable automerge_config.authenticated_web_messaging for deployment %s, so this configuration can never converge", deploymentId),
			errors.New("automerging authenticated web messaging requires authentication_settings with enabled = true on the web deployment's configuration"))
	}

	log.Printf("updated identity resolution for deployment %s", deploymentId)
	return readWebDeploymentIdentityResolution(ctx, d, meta)
}

func deleteWebDeploymentIdentityResolution(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getWebDeploymentIdentityResolutionProxy(sdkConfig)
	deploymentId := d.Id()

	log.Printf("resetting identity resolution for deployment %s to default", deploymentId)

	_, resp, getErr := proxy.getWebDeploymentIdentityResolution(ctx, deploymentId)
	if getErr != nil {
		if util.IsStatus404(resp) {
			log.Printf("parent deployment %s already deleted", deploymentId)
			return nil
		}
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to verify deployment %s before resetting identity resolution", deploymentId), resp)
	}

	defaultConfig := buildDefaultDeploymentIdentityResolutionConfig()
	_, putResp, putErr := proxy.putWebDeploymentIdentityResolution(ctx, deploymentId, defaultConfig)
	if putErr != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to reset identity resolution for deployment %s", deploymentId), putResp)
	}

	log.Printf("reset identity resolution for deployment %s to default", deploymentId)
	return nil
}
