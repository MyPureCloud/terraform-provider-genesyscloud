package conversations_messaging_integrations_instagram_identity_resolution

/*
The resource_genesyscloud_conversations_messaging_integrations_instagram_identity_resolution.go file contains all the methods that perform the core logic for the resource.
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

func getAllConversationsMessagingIntegrationsInstagramIdentityResolution(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getConversationsMessagingIntegrationsInstagramIdentityResolutionProxy(clientConfig)

	integrations, resp, err := proxy.instagramIntegrationProxy.GetAllConversationsMessagingIntegrationsInstagram(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, "failed to list instagram integrations for identity resolution export", resp)
	}
	if integrations == nil {
		return resources, nil
	}

	for _, integration := range *integrations {
		if integration.Id == nil || integration.Name == nil {
			continue
		}

		config, getResp, getErr := proxy.getConversationsMessagingIntegrationsInstagramIdentityResolution(ctx, *integration.Id)
		if getErr != nil {
			if util.IsStatus404(getResp) {
				continue
			}
			return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to get identity resolution for instagram integration %s", *integration.Id), getResp)
		}

		if isDefaultIdentityResolutionConfig(config) {
			continue
		}

		resources[*integration.Id] = &resourceExporter.ResourceMeta{BlockLabel: *integration.Name + "-identity-resolution"}
	}

	return resources, nil
}

func createConversationsMessagingIntegrationsInstagramIdentityResolution(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	instagramIntegrationId := d.Get("instagram_integration_id").(string)
	log.Printf("creating identity resolution for instagram integration %s", instagramIntegrationId)
	d.SetId(instagramIntegrationId)

	return updateConversationsMessagingIntegrationsInstagramIdentityResolution(ctx, d, meta)
}

func readConversationsMessagingIntegrationsInstagramIdentityResolution(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingIntegrationsInstagramIdentityResolutionProxy(sdkConfig)
	cc := consistencyChecker.NewConsistencyCheck(ctx, d, meta, ResourceConversationsMessagingIntegrationsInstagramIdentityResolution(), constants.ConsistencyChecks(), ResourceType)

	instagramIntegrationId := d.Id()
	log.Printf("reading identity resolution for instagram integration %s", instagramIntegrationId)

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		config, resp, getErr := proxy.getConversationsMessagingIntegrationsInstagramIdentityResolution(ctx, instagramIntegrationId)
		if getErr != nil {
			if util.IsStatus404(resp) {
				log.Printf("instagram integration %s not found, removing identity resolution from state", instagramIntegrationId)
				d.SetId("")
				return nil
			}
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read identity resolution for instagram integration %s: %s", instagramIntegrationId, getErr), resp))
		}

		_ = d.Set("instagram_integration_id", instagramIntegrationId)
		if config.ResolveIdentities != nil {
			_ = d.Set("resolve_identities", *config.ResolveIdentities)
		} else {
			_ = d.Set("resolve_identities", true)
		}
		if config.Division != nil && config.Division.Id != nil && !isUnassignedDivisionId(*config.Division.Id) {
			_ = d.Set("division_id", *config.Division.Id)
		}

		log.Printf("read identity resolution for instagram integration %s", instagramIntegrationId)
		return cc.CheckState(d)
	})
}

func updateConversationsMessagingIntegrationsInstagramIdentityResolution(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingIntegrationsInstagramIdentityResolutionProxy(sdkConfig)
	instagramIntegrationId := d.Id()

	config, err := buildIdentityResolutionInstagramConfig(d)
	if err != nil {
		return util.BuildDiagnosticError(ResourceType, "failed to build identity resolution config", err)
	}

	log.Printf("updating identity resolution for instagram integration %s", instagramIntegrationId)
	_, resp, putErr := proxy.putConversationsMessagingIntegrationsInstagramIdentityResolution(ctx, instagramIntegrationId, &config)
	if putErr != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to update identity resolution for instagram integration %s", instagramIntegrationId), resp)
	}

	log.Printf("updated identity resolution for instagram integration %s", instagramIntegrationId)
	return readConversationsMessagingIntegrationsInstagramIdentityResolution(ctx, d, meta)
}

func deleteConversationsMessagingIntegrationsInstagramIdentityResolution(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingIntegrationsInstagramIdentityResolutionProxy(sdkConfig)
	instagramIntegrationId := d.Id()

	log.Printf("resetting identity resolution for instagram integration %s to default", instagramIntegrationId)

	_, resp, getErr := proxy.getConversationsMessagingIntegrationsInstagramIdentityResolution(ctx, instagramIntegrationId)
	if getErr != nil {
		if util.IsStatus404(resp) {
			log.Printf("instagram integration %s not found, removing identity resolution from state", instagramIntegrationId)
			return nil
		}
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to verify instagram integration %s before resetting identity resolution", instagramIntegrationId), resp)
	}

	defaultConfig := buildDefaultIdentityResolutionInstagramConfig()
	_, putResp, putErr := proxy.putConversationsMessagingIntegrationsInstagramIdentityResolution(ctx, instagramIntegrationId, &defaultConfig)
	if putErr != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to reset identity resolution for instagram integration %s", instagramIntegrationId), putResp)
	}

	log.Printf("reset identity resolution for instagram integration %s to default", instagramIntegrationId)
	return nil
}
