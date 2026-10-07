package conversations_messaging_integrations_apple_identity_resolution

/*
The resource_genesyscloud_conversations_messaging_integrations_apple_identity_resolution.go file contains all the methods that perform the core logic for the resource.
*/

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v199/platformclientv2"
	consistencyChecker "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
)

func getAllConversationsMessagingIntegrationsAppleIdentityResolution(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getConversationsMessagingIntegrationsAppleIdentityResolutionProxy(clientConfig)

	integrations, resp, err := proxy.appleIntegrationProxy.GetAllConversationsMessagingIntegrationsApple(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, "failed to list apple integrations for identity resolution export", resp)
	}
	if integrations == nil {
		return resources, nil
	}

	for _, integration := range *integrations {
		if integration.Id == nil || integration.Name == nil {
			continue
		}

		config, getResp, getErr := proxy.getConversationsMessagingIntegrationsAppleIdentityResolution(ctx, *integration.Id)
		if getErr != nil {
			if util.IsStatus404(getResp) {
				continue
			}
			return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to get identity resolution for apple integration %s", *integration.Id), getResp)
		}

		if isDefaultIdentityResolutionConfig(config) {
			continue
		}

		resources[*integration.Id] = &resourceExporter.ResourceMeta{BlockLabel: *integration.Name + "-identity-resolution"}
	}

	return resources, nil
}

func createConversationsMessagingIntegrationsAppleIdentityResolution(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	appleIntegrationId := d.Get("apple_integration_id").(string)
	log.Printf("creating identity resolution for apple integration %s", appleIntegrationId)
	d.SetId(appleIntegrationId)

	return updateConversationsMessagingIntegrationsAppleIdentityResolution(ctx, d, meta)
}

func readConversationsMessagingIntegrationsAppleIdentityResolution(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingIntegrationsAppleIdentityResolutionProxy(sdkConfig)
	cc := consistencyChecker.NewConsistencyCheck(ctx, d, meta, ResourceConversationsMessagingIntegrationsAppleIdentityResolution(), constants.ConsistencyChecks(), ResourceType)

	appleIntegrationId := d.Id()
	log.Printf("reading identity resolution for apple integration %s", appleIntegrationId)

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		config, resp, getErr := proxy.getConversationsMessagingIntegrationsAppleIdentityResolution(ctx, appleIntegrationId)
		if getErr != nil {
			if util.IsStatus404(resp) {
				log.Printf("apple integration %s not found, removing identity resolution from state", appleIntegrationId)
				d.SetId("")
				return nil
			}
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read identity resolution for apple integration %s: %s", appleIntegrationId, getErr), resp))
		}

		_ = d.Set("apple_integration_id", appleIntegrationId)
		if config.ResolveIdentities != nil {
			_ = d.Set("resolve_identities", *config.ResolveIdentities)
		} else {
			_ = d.Set("resolve_identities", true)
		}
		if config.Division != nil && config.Division.Id != nil && !isUnassignedDivisionId(*config.Division.Id) {
			_ = d.Set("division_id", *config.Division.Id)
		}

		log.Printf("read identity resolution for apple integration %s", appleIntegrationId)
		return cc.CheckState(d)
	})
}

func updateConversationsMessagingIntegrationsAppleIdentityResolution(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingIntegrationsAppleIdentityResolutionProxy(sdkConfig)
	appleIntegrationId := d.Id()

	config, err := buildIdentityResolutionAppleConfig(d)
	if err != nil {
		return util.BuildDiagnosticError(ResourceType, "failed to build identity resolution config", err)
	}

	log.Printf("updating identity resolution for apple integration %s", appleIntegrationId)
	_, resp, putErr := proxy.putConversationsMessagingIntegrationsAppleIdentityResolution(ctx, appleIntegrationId, &config)
	if putErr != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to update identity resolution for apple integration %s", appleIntegrationId), resp)
	}

	log.Printf("updated identity resolution for apple integration %s", appleIntegrationId)
	return readConversationsMessagingIntegrationsAppleIdentityResolution(ctx, d, meta)
}

func deleteConversationsMessagingIntegrationsAppleIdentityResolution(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingIntegrationsAppleIdentityResolutionProxy(sdkConfig)
	appleIntegrationId := d.Id()

	log.Printf("resetting identity resolution for apple integration %s to default", appleIntegrationId)

	_, resp, getErr := proxy.getConversationsMessagingIntegrationsAppleIdentityResolution(ctx, appleIntegrationId)
	if getErr != nil {
		if util.IsStatus404(resp) {
			log.Printf("apple integration %s not found, removing identity resolution from state", appleIntegrationId)
			return nil
		}
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to verify apple integration %s before resetting identity resolution", appleIntegrationId), resp)
	}

	defaultConfig := buildDefaultIdentityResolutionAppleConfig()
	_, putResp, putErr := proxy.putConversationsMessagingIntegrationsAppleIdentityResolution(ctx, appleIntegrationId, &defaultConfig)
	if putErr != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to reset identity resolution for apple integration %s", appleIntegrationId), putResp)
	}

	log.Printf("reset identity resolution for apple integration %s to default", appleIntegrationId)
	return nil
}
