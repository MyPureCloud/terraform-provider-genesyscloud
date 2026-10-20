package conversations_messaging_integrations_open_identity_resolution

import (
	"context"

	"github.com/mypurecloud/platform-client-sdk-go/v199/platformclientv2"
	open "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_integrations_open"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
)

/*
The genesyscloud_conversations_messaging_integrations_open_identity_resolution_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

var internalProxy *conversationsMessagingIntegrationsOpenIdentityResolutionProxy

type getConversationsMessagingIntegrationsOpenIdentityResolutionFunc func(ctx context.Context, p *conversationsMessagingIntegrationsOpenIdentityResolutionProxy, openIntegrationId string) (openIdentityResolution *platformclientv2.Openmessagingidentityresolutionconfig, response *platformclientv2.APIResponse, err error)
type putConversationsMessagingIntegrationsOpenIdentityResolutionFunc func(ctx context.Context, p *conversationsMessagingIntegrationsOpenIdentityResolutionProxy, openIntegrationId string, config *platformclientv2.Openmessagingidentityresolutionconfig) (openIdentityResolution *platformclientv2.Openmessagingidentityresolutionconfig, response *platformclientv2.APIResponse, err error)
type getConversationsMessagingIntegrationsOpenByIdFunc func(ctx context.Context, p *conversationsMessagingIntegrationsOpenIdentityResolutionProxy, openIntegrationId string) (openIntegration *platformclientv2.Openintegration, response *platformclientv2.APIResponse, err error)

type conversationsMessagingIntegrationsOpenIdentityResolutionProxy struct {
	clientConfig                                                    *platformclientv2.Configuration
	conversationsApi                                                *platformclientv2.ConversationsApi
	getConversationsMessagingIntegrationsOpenIdentityResolutionAttr getConversationsMessagingIntegrationsOpenIdentityResolutionFunc
	putConversationsMessagingIntegrationsOpenIdentityResolutionAttr putConversationsMessagingIntegrationsOpenIdentityResolutionFunc
	getConversationsMessagingIntegrationsOpenByIdAttr               getConversationsMessagingIntegrationsOpenByIdFunc
	openIntegrationProxy                                            *open.ConversationsMessagingIntegrationsOpenProxy
}

func newConversationsMessagingIntegrationsOpenIdentityResolutionProxy(clientConfig *platformclientv2.Configuration) *conversationsMessagingIntegrationsOpenIdentityResolutionProxy {
	api := platformclientv2.NewConversationsApiWithConfig(clientConfig)
	openIntegrationProxt := open.GetConversationsMessagingIntegrationsOpenProxy(clientConfig)

	return &conversationsMessagingIntegrationsOpenIdentityResolutionProxy{
		clientConfig:     clientConfig,
		conversationsApi: api,
		getConversationsMessagingIntegrationsOpenIdentityResolutionAttr: getConversationsMessagingIntegrationsOpenIdentityResolutionFn,
		putConversationsMessagingIntegrationsOpenIdentityResolutionAttr: putConversationsMessagingIntegrationsOpenIdentityResolutionFn,
		getConversationsMessagingIntegrationsOpenByIdAttr:               getConversationsMessagingIntegrationsOpenByIdFn,
		openIntegrationProxy: openIntegrationProxt,
	}
}

func getConversationsMessagingIntegrationsOpenIdentityResolutionProxy(clientConfig *platformclientv2.Configuration) *conversationsMessagingIntegrationsOpenIdentityResolutionProxy {
	if internalProxy == nil {
		internalProxy = newConversationsMessagingIntegrationsOpenIdentityResolutionProxy(clientConfig)
	}
	return internalProxy
}

func (p *conversationsMessagingIntegrationsOpenIdentityResolutionProxy) getConversationsMessagingIntegrationsOpenIdentityResolution(ctx context.Context, openIntegrationId string) (openIdentityResolution *platformclientv2.Openmessagingidentityresolutionconfig, response *platformclientv2.APIResponse, err error) {
	return p.getConversationsMessagingIntegrationsOpenIdentityResolutionAttr(ctx, p, openIntegrationId)
}

func (p *conversationsMessagingIntegrationsOpenIdentityResolutionProxy) putConversationsMessagingIntegrationsOpenIdentityResolution(ctx context.Context, openIntegrationId string, config *platformclientv2.Openmessagingidentityresolutionconfig) (openIdentityResolution *platformclientv2.Openmessagingidentityresolutionconfig, response *platformclientv2.APIResponse, err error) {
	return p.putConversationsMessagingIntegrationsOpenIdentityResolutionAttr(ctx, p, openIntegrationId, config)
}

func (p *conversationsMessagingIntegrationsOpenIdentityResolutionProxy) getConversationsMessagingIntegrationsOpenById(ctx context.Context, openIntegrationId string) (openIntegration *platformclientv2.Openintegration, response *platformclientv2.APIResponse, err error) {
	return p.getConversationsMessagingIntegrationsOpenByIdAttr(ctx, p, openIntegrationId)
}

func getConversationsMessagingIntegrationsOpenIdentityResolutionFn(ctx context.Context, p *conversationsMessagingIntegrationsOpenIdentityResolutionProxy, openIntegrationId string) (openIdentityResolution *platformclientv2.Openmessagingidentityresolutionconfig, response *platformclientv2.APIResponse, err error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	return p.conversationsApi.GetConversationsMessagingIdentityresolutionIntegrationsOpenIntegrationId(openIntegrationId)
}

func putConversationsMessagingIntegrationsOpenIdentityResolutionFn(ctx context.Context, p *conversationsMessagingIntegrationsOpenIdentityResolutionProxy, openIntegrationId string, config *platformclientv2.Openmessagingidentityresolutionconfig) (openIdentityResolution *platformclientv2.Openmessagingidentityresolutionconfig, response *platformclientv2.APIResponse, err error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	return p.conversationsApi.PutConversationsMessagingIdentityresolutionIntegrationsOpenIntegrationId(openIntegrationId, *config)
}

func getConversationsMessagingIntegrationsOpenByIdFn(ctx context.Context, p *conversationsMessagingIntegrationsOpenIdentityResolutionProxy, openIntegrationId string) (openIntegration *platformclientv2.Openintegration, response *platformclientv2.APIResponse, err error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	return p.openIntegrationProxy.GetConversationsMessagingIntegrationsOpenById(ctx, openIntegrationId)
}
