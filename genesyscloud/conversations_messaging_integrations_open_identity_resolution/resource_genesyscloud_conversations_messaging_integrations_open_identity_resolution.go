package conversations_messaging_integrations_open_identity_resolution

/*
The resource_genesyscloud_conversations_messaging_integrations_open_identity_resolution.go file contains all the methods that perform the core logic for the resource.
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

func getAllConversationsMessagingIntegrationsOpenIdentityResolution(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getConversationsMessagingIntegrationsOpenIdentityResolutionProxy(clientConfig)

	integrations, resp, err := proxy.openIntegrationProxy.GetAllConversationsMessagingIntegrationsOpen(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, "failed to list open integrations for identity resolution export", resp)
	}
	if integrations == nil {
		return resources, nil
	}

	for _, integration := range *integrations {
		if integration.Id == nil || integration.Name == nil {
			continue
		}

		config, getResp, getErr := proxy.getConversationsMessagingIntegrationsOpenIdentityResolution(ctx, *integration.Id)
		if getErr != nil {
			if util.IsStatus404(getResp) {
				continue
			}
			return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to get identity resolution for open integration %s", *integration.Id), getResp)
		}

		if isDefaultIdentityResolutionConfig(config) {
			continue
		}

		resources[*integration.Id] = &resourceExporter.ResourceMeta{BlockLabel: *integration.Name + "-identity-resolution"}
	}

	return resources, nil
}

func createConversationsMessagingIntegrationsOpenIdentityResolution(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	openIntegrationId := d.Get("open_integration_id").(string)
	log.Printf("creating identity resolution for open integration %s", openIntegrationId)
	d.SetId(openIntegrationId)

	return updateConversationsMessagingIntegrationsOpenIdentityResolution(ctx, d, meta)
}

func readConversationsMessagingIntegrationsOpenIdentityResolution(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingIntegrationsOpenIdentityResolutionProxy(sdkConfig)
	cc := consistencyChecker.NewConsistencyCheck(ctx, d, meta, ResourceConversationsMessagingIntegrationsOpenIdentityResolution(), constants.ConsistencyChecks(), ResourceType)

	openIntegrationId := d.Id()
	log.Printf("reading identity resolution for open integration %s", openIntegrationId)

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		config, resp, getErr := proxy.getConversationsMessagingIntegrationsOpenIdentityResolution(ctx, openIntegrationId)
		if getErr != nil {
			if util.IsStatus404(resp) {
				log.Printf("open integration %s not found, removing identity resolution from state", openIntegrationId)
				d.SetId("")
				return nil
			}
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read identity resolution for open integration %s: %s", openIntegrationId, getErr), resp))
		}

		_ = d.Set("open_integration_id", openIntegrationId)
		if config.ResolveIdentities != nil {
			_ = d.Set("resolve_identities", *config.ResolveIdentities)
		} else {
			_ = d.Set("resolve_identities", true)
		}
		if config.Division != nil && config.Division.Id != nil && !isUnassignedDivisionId(*config.Division.Id) {
			_ = d.Set("division_id", *config.Division.Id)
		}
		if config.ExternalSource != nil && config.ExternalSource.Id != nil {
			_ = d.Set("external_source_id", *config.ExternalSource.Id)
		}

		log.Printf("read identity resolution for open integration %s", openIntegrationId)
		return cc.CheckState(d)
	})
}

func updateConversationsMessagingIntegrationsOpenIdentityResolution(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingIntegrationsOpenIdentityResolutionProxy(sdkConfig)
	openIntegrationId := d.Id()

	config, err := buildIdentityResolutionOpenConfig(d)
	if err != nil {
		return util.BuildDiagnosticError(ResourceType, "failed to build identity resolution config", err)
	}

	log.Printf("updating identity resolution for open integration %s", openIntegrationId)
	_, resp, putErr := proxy.putConversationsMessagingIntegrationsOpenIdentityResolution(ctx, openIntegrationId, &config)
	if putErr != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to update identity resolution for open integration %s", openIntegrationId), resp)
	}

	log.Printf("updated identity resolution for open integration %s", openIntegrationId)
	return readConversationsMessagingIntegrationsOpenIdentityResolution(ctx, d, meta)
}

func deleteConversationsMessagingIntegrationsOpenIdentityResolution(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingIntegrationsOpenIdentityResolutionProxy(sdkConfig)
	openIntegrationId := d.Id()

	log.Printf("resetting identity resolution for open integration %s to default", openIntegrationId)

	_, resp, getErr := proxy.getConversationsMessagingIntegrationsOpenIdentityResolution(ctx, openIntegrationId)
	if getErr != nil {
		if util.IsStatus404(resp) {
			log.Printf("open integration %s not found, removing identity resolution from state", openIntegrationId)
			return nil
		}
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to verify open integration %s before resetting identity resolution", openIntegrationId), resp)
	}

	defaultConfig := buildDefaultIdentityResolutionOpenConfig()
	_, putResp, putErr := proxy.putConversationsMessagingIntegrationsOpenIdentityResolution(ctx, openIntegrationId, &defaultConfig)
	if putErr != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to reset identity resolution for open integration %s", openIntegrationId), putResp)
	}

	log.Printf("reset identity resolution for open integration %s to default", openIntegrationId)
	return nil
}
