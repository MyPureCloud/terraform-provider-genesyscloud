package conversations_messaging_integrations_whatsapp_identity_resolution

/*
The resource_genesyscloud_conversations_messaging_integrations_whatsapp_identity_resolution.go file contains all the methods that perform the core logic for the resource.
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

func getAllConversationsMessagingIntegrationsWhatsappIdentityResolution(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getConversationsMessagingIntegrationsWhatsappIdentityResolutionProxy(clientConfig)

	integrations, resp, err := proxy.whatsappIntegrationProxy.GetAllConversationsMessagingIntegrationsWhatsapp(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, "failed to list whatsapp integrations for identity resolution export", resp)
	}
	if integrations == nil {
		return resources, nil
	}

	for _, integration := range *integrations {
		if integration.Id == nil || integration.Name == nil {
			continue
		}

		config, getResp, getErr := proxy.getConversationsMessagingIntegrationsWhatsappIdentityResolution(ctx, *integration.Id)
		if getErr != nil {
			if util.IsStatus404(getResp) {
				continue
			}
			return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to get identity resolution for whatsapp integration %s", *integration.Id), getResp)
		}

		if isDefaultIdentityResolutionConfig(config) {
			continue
		}

		resources[*integration.Id] = &resourceExporter.ResourceMeta{BlockLabel: *integration.Name + "-identity-resolution"}
	}

	return resources, nil
}

func createConversationsMessagingIntegrationsWhatsappIdentityResolution(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	whatsappIntegrationId := d.Get("whatsapp_integration_id").(string)
	log.Printf("creating identity resolution for whatsapp integration %s", whatsappIntegrationId)
	d.SetId(whatsappIntegrationId)

	return updateConversationsMessagingIntegrationsWhatsappIdentityResolution(ctx, d, meta)
}

func readConversationsMessagingIntegrationsWhatsappIdentityResolution(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingIntegrationsWhatsappIdentityResolutionProxy(sdkConfig)
	cc := consistencyChecker.NewConsistencyCheck(ctx, d, meta, ResourceConversationsMessagingIntegrationsWhatsappIdentityResolution(), constants.ConsistencyChecks(), ResourceType)

	whatsappIntegrationId := d.Id()
	log.Printf("reading identity resolution for whatsapp integration %s", whatsappIntegrationId)

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		config, resp, getErr := proxy.getConversationsMessagingIntegrationsWhatsappIdentityResolution(ctx, whatsappIntegrationId)
		if getErr != nil {
			if util.IsStatus404(resp) {
				log.Printf("whatsapp integration %s not found, removing identity resolution from state", whatsappIntegrationId)
				d.SetId("")
				return nil
			}
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read identity resolution for whatsapp integration %s: %s", whatsappIntegrationId, getErr), resp))
		}

		_ = d.Set("whatsapp_integration_id", whatsappIntegrationId)
		if config.ResolveIdentities != nil {
			_ = d.Set("resolve_identities", *config.ResolveIdentities)
		} else {
			_ = d.Set("resolve_identities", true)
		}
		if config.Division != nil && config.Division.Id != nil && !isUnassignedDivisionId(*config.Division.Id) {
			_ = d.Set("division_id", *config.Division.Id)
		}

		log.Printf("read identity resolution for whatsapp integration %s", whatsappIntegrationId)
		return cc.CheckState(d)
	})
}

func updateConversationsMessagingIntegrationsWhatsappIdentityResolution(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingIntegrationsWhatsappIdentityResolutionProxy(sdkConfig)
	whatsappIntegrationId := d.Id()

	config, err := buildIdentityResolutionWhatsappConfig(d)
	if err != nil {
		return util.BuildDiagnosticError(ResourceType, "failed to build identity resolution config", err)
	}

	log.Printf("updating identity resolution for whatsapp integration %s", whatsappIntegrationId)
	_, resp, putErr := proxy.putConversationsMessagingIntegrationsWhatsappIdentityResolution(ctx, whatsappIntegrationId, &config)
	if putErr != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to update identity resolution for whatsapp integration %s", whatsappIntegrationId), resp)
	}

	log.Printf("updated identity resolution for whatsapp integration %s", whatsappIntegrationId)
	return readConversationsMessagingIntegrationsWhatsappIdentityResolution(ctx, d, meta)
}

func deleteConversationsMessagingIntegrationsWhatsappIdentityResolution(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingIntegrationsWhatsappIdentityResolutionProxy(sdkConfig)
	whatsappIntegrationId := d.Id()

	log.Printf("resetting identity resolution for whatsapp integration %s to default", whatsappIntegrationId)

	_, resp, getErr := proxy.getConversationsMessagingIntegrationsWhatsappIdentityResolution(ctx, whatsappIntegrationId)
	if getErr != nil {
		if util.IsStatus404(resp) {
			log.Printf("whatsapp integration %s not found, removing identity resolution from state", whatsappIntegrationId)
			return nil
		}
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to verify whatsapp integration %s before resetting identity resolution", whatsappIntegrationId), resp)
	}

	defaultConfig := buildDefaultIdentityResolutionWhatsappConfig()
	_, putResp, putErr := proxy.putConversationsMessagingIntegrationsWhatsappIdentityResolution(ctx, whatsappIntegrationId, &defaultConfig)
	if putErr != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to reset identity resolution for whatsapp integration %s", whatsappIntegrationId), putResp)
	}

	log.Printf("reset identity resolution for whatsapp integration %s to default", whatsappIntegrationId)
	return nil
}
