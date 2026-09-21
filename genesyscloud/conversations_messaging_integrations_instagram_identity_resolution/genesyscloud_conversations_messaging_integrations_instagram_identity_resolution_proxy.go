package conversations_messaging_integrations_instagram_identity_resolution

import (
	"context"

	"github.com/mypurecloud/platform-client-sdk-go/v199/platformclientv2"
	instagram "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_integrations_instagram"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
)

/*
The genesyscloud_conversations_messaging_integrations_instagram_identity_resolution_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

var internalProxy *conversationsMessagingIntegrationsInstagramIdentityResolutionProxy

type getConversationsMessagingIntegrationsInstagramIdentityResolutionFunc func(ctx context.Context, p *conversationsMessagingIntegrationsInstagramIdentityResolutionProxy, instagramIntegrationId string) (instagramIdentityResolution *platformclientv2.Instagramidentityresolutionconfig, response *platformclientv2.APIResponse, err error)
type putConversationsMessagingIntegrationsInstagramIdentityResolutionFunc func(ctx context.Context, p *conversationsMessagingIntegrationsInstagramIdentityResolutionProxy, instagramIntegrationId string, config *platformclientv2.Instagramidentityresolutionconfig) (instagramIdentityResolution *platformclientv2.Instagramidentityresolutionconfig, response *platformclientv2.APIResponse, err error)
type getConversationsMessagingIntegrationsInstagramByIdFunc func(ctx context.Context, p *conversationsMessagingIntegrationsInstagramIdentityResolutionProxy, instagramIntegrationId string) (instagramIntegration *platformclientv2.Instagramintegration, response *platformclientv2.APIResponse, err error)

type conversationsMessagingIntegrationsInstagramIdentityResolutionProxy struct {
	clientConfig                                                         *platformclientv2.Configuration
	conversationsApi                                                     *platformclientv2.ConversationsApi
	getConversationsMessagingIntegrationsInstagramIdentityResolutionAttr getConversationsMessagingIntegrationsInstagramIdentityResolutionFunc
	putConversationsMessagingIntegrationsInstagramIdentityResolutionAttr putConversationsMessagingIntegrationsInstagramIdentityResolutionFunc
	getConversationsMessagingIntegrationsInstagramByIdAttr               getConversationsMessagingIntegrationsInstagramByIdFunc
	instagramIntegrationProxy                                            *instagram.ConversationsMessagingIntegrationsInstagramProxy
}

func newConversationsMessagingIntegrationsInstagramIdentityResolutionProxy(clientConfig *platformclientv2.Configuration) *conversationsMessagingIntegrationsInstagramIdentityResolutionProxy {
	api := platformclientv2.NewConversationsApiWithConfig(clientConfig)
	instagramIntegrationProxt := instagram.GetConversationsMessagingIntegrationsInstagramProxy(clientConfig)

	return &conversationsMessagingIntegrationsInstagramIdentityResolutionProxy{
		clientConfig:     clientConfig,
		conversationsApi: api,
		getConversationsMessagingIntegrationsInstagramIdentityResolutionAttr: getConversationsMessagingIntegrationsInstagramIdentityResolutionFn,
		putConversationsMessagingIntegrationsInstagramIdentityResolutionAttr: putConversationsMessagingIntegrationsInstagramIdentityResolutionFn,
		getConversationsMessagingIntegrationsInstagramByIdAttr:               getConversationsMessagingIntegrationsInstagramByIdFn,
		instagramIntegrationProxy: instagramIntegrationProxt,
	}
}

func getConversationsMessagingIntegrationsInstagramIdentityResolutionProxy(clientConfig *platformclientv2.Configuration) *conversationsMessagingIntegrationsInstagramIdentityResolutionProxy {
	if internalProxy == nil {
		internalProxy = newConversationsMessagingIntegrationsInstagramIdentityResolutionProxy(clientConfig)
	}
	return internalProxy
}

func (p *conversationsMessagingIntegrationsInstagramIdentityResolutionProxy) getConversationsMessagingIntegrationsInstagramIdentityResolution(ctx context.Context, instagramIntegrationId string) (instagramIdentityResolution *platformclientv2.Instagramidentityresolutionconfig, response *platformclientv2.APIResponse, err error) {
	return p.getConversationsMessagingIntegrationsInstagramIdentityResolutionAttr(ctx, p, instagramIntegrationId)
}

func (p *conversationsMessagingIntegrationsInstagramIdentityResolutionProxy) putConversationsMessagingIntegrationsInstagramIdentityResolution(ctx context.Context, instagramIntegrationId string, config *platformclientv2.Instagramidentityresolutionconfig) (instagramIdentityResolution *platformclientv2.Instagramidentityresolutionconfig, response *platformclientv2.APIResponse, err error) {
	return p.putConversationsMessagingIntegrationsInstagramIdentityResolutionAttr(ctx, p, instagramIntegrationId, config)
}

func (p *conversationsMessagingIntegrationsInstagramIdentityResolutionProxy) getConversationsMessagingIntegrationsInstagramById(ctx context.Context, instagramIntegrationId string) (instagramIntegration *platformclientv2.Instagramintegration, response *platformclientv2.APIResponse, err error) {
	return p.getConversationsMessagingIntegrationsInstagramByIdAttr(ctx, p, instagramIntegrationId)
}

func getConversationsMessagingIntegrationsInstagramIdentityResolutionFn(ctx context.Context, p *conversationsMessagingIntegrationsInstagramIdentityResolutionProxy, instagramIntegrationId string) (instagramIdentityResolution *platformclientv2.Instagramidentityresolutionconfig, response *platformclientv2.APIResponse, err error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	return p.conversationsApi.GetConversationsMessagingIntegrationsInstagramIdentityresolution(instagramIntegrationId)
}

func putConversationsMessagingIntegrationsInstagramIdentityResolutionFn(ctx context.Context, p *conversationsMessagingIntegrationsInstagramIdentityResolutionProxy, instagramIntegrationId string, config *platformclientv2.Instagramidentityresolutionconfig) (instagramIdentityResolution *platformclientv2.Instagramidentityresolutionconfig, response *platformclientv2.APIResponse, err error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	return p.conversationsApi.PutConversationsMessagingIntegrationsInstagramIdentityresolution(instagramIntegrationId, *config)
}

func getConversationsMessagingIntegrationsInstagramByIdFn(ctx context.Context, p *conversationsMessagingIntegrationsInstagramIdentityResolutionProxy, instagramIntegrationId string) (instagramIntegration *platformclientv2.Instagramintegration, response *platformclientv2.APIResponse, err error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	return p.instagramIntegrationProxy.GetConversationsMessagingIntegrationsInstagramById(ctx, instagramIntegrationId)
}
