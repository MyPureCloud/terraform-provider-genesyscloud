package integration_config

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"

	consistencyChecker "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	featureToggles "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/feature_toggles"
)

// getAllIntegrationConfigs returns all integration configs for export
func getAllIntegrationConfigs(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)

	if !featureToggles.ICToggleExists() {
		log.Printf("Environment variable %s not set, skipping exporter for %s", featureToggles.ICToggleName(), ResourceType)
		return nil, nil
	}

	// Reuse the integrations API to list all integrations that have config
	api := platformclientv2.NewIntegrationsApiWithConfig(clientConfig)

	const pageSize = 100
	for pageNum := 1; ; pageNum++ {
		integrations, resp, err := api.GetIntegrations(pageSize, pageNum, "", nil, "", "", nil, "", "", "")
		if err != nil {
			return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get integrations: %s", err), resp)
		}

		if integrations.Entities == nil || len(*integrations.Entities) == 0 {
			break
		}

		for _, integration := range *integrations.Entities {
			if integration.Id != nil && integration.Name != nil {
				resources[*integration.Id+"/config"] = &resourceExporter.ResourceMeta{BlockLabel: *integration.Name + "_config"}
			}
		}

		if integrations.NextUri == nil || *integrations.NextUri == "" {
			break
		}
	}

	return resources, nil
}

// createIntegrationConfig creates the integration config (actually a PUT since config always exists with the integration)
func createIntegrationConfig(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if !featureToggles.ICToggleExists() {
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Environment variable %s not set", featureToggles.ICToggleName()),
			fmt.Errorf("set environment variable %s to use this resource", featureToggles.ICToggleName()))
	}

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getIntegrationConfigProxy(sdkConfig)

	integrationId := d.Get("integration_id").(string)
	log.Printf("Creating integration config for integration %s", integrationId)

	// Get current config to obtain version number (required for PUT)
	currentConfig, resp, err := proxy.getConfig(ctx, integrationId)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get current config for integration %s: %s", integrationId, err), resp)
	}

	// Build and update
	config := buildIntegrationConfig(d, currentConfig.Version)
	_, resp, err = proxy.updateConfig(ctx, integrationId, config)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create config for integration %s: %s", integrationId, err), resp)
	}

	d.SetId(integrationId + "/config")
	log.Printf("Created integration config for integration %s", integrationId)

	return readIntegrationConfig(ctx, d, meta)
}

// readIntegrationConfig reads the integration config from the API
func readIntegrationConfig(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if !featureToggles.ICToggleExists() {
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Environment variable %s not set", featureToggles.ICToggleName()),
			fmt.Errorf("set environment variable %s to use this resource", featureToggles.ICToggleName()))
	}

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getIntegrationConfigProxy(sdkConfig)
	cc := consistencyChecker.NewConsistencyCheck(ctx, d, meta, ResourceIntegrationConfig(), constants.ConsistencyChecks(), ResourceType)

	integrationId := strings.Split(d.Id(), "/")[0]

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		log.Printf("Reading integration config for integration %s", integrationId)

		config, resp, err := proxy.getConfig(ctx, integrationId)
		if err != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read config for integration %s: %s", integrationId, err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read config for integration %s: %s", integrationId, err), resp))
		}

		_ = d.Set("integration_id", integrationId)
		flattenIntegrationConfig(d, config)

		log.Printf("Read integration config for integration %s", integrationId)
		return cc.CheckState(d)
	})
}

// updateIntegrationConfig updates the integration config
func updateIntegrationConfig(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if !featureToggles.ICToggleExists() {
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Environment variable %s not set", featureToggles.ICToggleName()),
			fmt.Errorf("set environment variable %s to use this resource", featureToggles.ICToggleName()))
	}

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getIntegrationConfigProxy(sdkConfig)

	integrationId := strings.Split(d.Id(), "/")[0]
	log.Printf("Updating integration config for integration %s", integrationId)

	// Retry on version mismatch (optimistic locking)
	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get latest version
		currentConfig, resp, err := proxy.getConfig(ctx, integrationId)
		if err != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get current config for integration %s: %s", integrationId, err), resp)
		}

		config := buildIntegrationConfig(d, currentConfig.Version)
		_, resp, err = proxy.updateConfig(ctx, integrationId, config)
		if err != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update config for integration %s: %s", integrationId, err), resp)
		}
		return nil, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated integration config for integration %s", integrationId)
	return readIntegrationConfig(ctx, d, meta)
}

// deleteIntegrationConfig clears the integration config (sets credentials to empty)
func deleteIntegrationConfig(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getIntegrationConfigProxy(sdkConfig)

	integrationId := strings.Split(d.Id(), "/")[0]
	log.Printf("Deleting (clearing) integration config for integration %s", integrationId)

	// Get current config for version
	currentConfig, resp, err := proxy.getConfig(ctx, integrationId)
	if err != nil {
		if util.IsStatus404(resp) {
			log.Printf("Integration %s already deleted, skipping config cleanup", integrationId)
			return nil
		}
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get config for integration %s before delete: %s", integrationId, err), resp)
	}

	// Clear credentials and properties
	emptyCredentials := make(map[string]platformclientv2.Credentialinfo)
	emptyNotes := ""
	var emptyProps interface{} = map[string]interface{}{}
	var emptyAdv interface{} = map[string]interface{}{}
	clearConfig := &platformclientv2.Integrationconfiguration{
		Version:     currentConfig.Version,
		Name:        currentConfig.Name, // Keep the name
		Notes:       &emptyNotes,
		Properties:  &emptyProps,
		Advanced:    &emptyAdv,
		Credentials: &emptyCredentials,
	}

	_, resp, err = proxy.updateConfig(ctx, integrationId, clearConfig)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to clear config for integration %s: %s", integrationId, err), resp)
	}

	log.Printf("Deleted (cleared) integration config for integration %s", integrationId)
	return nil
}
