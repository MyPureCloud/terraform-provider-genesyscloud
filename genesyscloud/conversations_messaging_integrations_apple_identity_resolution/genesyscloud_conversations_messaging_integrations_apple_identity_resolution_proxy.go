package conversations_messaging_integrations_apple_identity_resolution

import (
	"context"

	"github.com/mypurecloud/platform-client-sdk-go/v199/platformclientv2"
	apple "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_integrations_apple"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
)

/*
The genesyscloud_conversations_messaging_integrations_apple_identity_resolution_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

var internalProxy *conversationsMessagingIntegrationsAppleIdentityResolutionProxy

type getConversationsMessagingIntegrationsAppleIdentityResolutionFunc func(ctx context.Context, p *conversationsMessagingIntegrationsAppleIdentityResolutionProxy, appleIntegrationId string) (appleIdentityResolution *platformclientv2.Appleidentityresolutionconfig, response *platformclientv2.APIResponse, err error)
type putConversationsMessagingIntegrationsAppleIdentityResolutionFunc func(ctx context.Context, p *conversationsMessagingIntegrationsAppleIdentityResolutionProxy, appleIntegrationId string, config *platformclientv2.Appleidentityresolutionconfig) (appleIdentityResolution *platformclientv2.Appleidentityresolutionconfig, response *platformclientv2.APIResponse, err error)
type getConversationsMessagingIntegrationsAppleByIdFunc func(ctx context.Context, p *conversationsMessagingIntegrationsAppleIdentityResolutionProxy, appleIntegrationId string) (appleIntegration *platformclientv2.Appleintegration, response *platformclientv2.APIResponse, err error)

type conversationsMessagingIntegrationsAppleIdentityResolutionProxy struct {
	clientConfig                                                     *platformclientv2.Configuration
	conversationsApi                                                 *platformclientv2.ConversationsApi
	getConversationsMessagingIntegrationsAppleIdentityResolutionAttr getConversationsMessagingIntegrationsAppleIdentityResolutionFunc
	putConversationsMessagingIntegrationsAppleIdentityResolutionAttr putConversationsMessagingIntegrationsAppleIdentityResolutionFunc
	getConversationsMessagingIntegrationsAppleByIdAttr               getConversationsMessagingIntegrationsAppleByIdFunc
	appleIntegrationProxy                                            *apple.ConversationsMessagingIntegrationsAppleProxy
}

func newConversationsMessagingIntegrationsAppleIdentityResolutionProxy(clientConfig *platformclientv2.Configuration) *conversationsMessagingIntegrationsAppleIdentityResolutionProxy {
	api := platformclientv2.NewConversationsApiWithConfig(clientConfig)
	appleIntegrationProxt := apple.GetConversationsMessagingIntegrationsAppleProxy(clientConfig)

	return &conversationsMessagingIntegrationsAppleIdentityResolutionProxy{
		clientConfig:     clientConfig,
		conversationsApi: api,
		getConversationsMessagingIntegrationsAppleIdentityResolutionAttr: getConversationsMessagingIntegrationsAppleIdentityResolutionFn,
		putConversationsMessagingIntegrationsAppleIdentityResolutionAttr: putConversationsMessagingIntegrationsAppleIdentityResolutionFn,
		getConversationsMessagingIntegrationsAppleByIdAttr:               getConversationsMessagingIntegrationsAppleByIdFn,
		appleIntegrationProxy: appleIntegrationProxt,
	}
}

func getConversationsMessagingIntegrationsAppleIdentityResolutionProxy(clientConfig *platformclientv2.Configuration) *conversationsMessagingIntegrationsAppleIdentityResolutionProxy {
	if internalProxy == nil {
		internalProxy = newConversationsMessagingIntegrationsAppleIdentityResolutionProxy(clientConfig)
	}
	return internalProxy
}

func (p *conversationsMessagingIntegrationsAppleIdentityResolutionProxy) getConversationsMessagingIntegrationsAppleIdentityResolution(ctx context.Context, appleIntegrationId string) (appleIdentityResolution *platformclientv2.Appleidentityresolutionconfig, response *platformclientv2.APIResponse, err error) {
	return p.getConversationsMessagingIntegrationsAppleIdentityResolutionAttr(ctx, p, appleIntegrationId)
}

func (p *conversationsMessagingIntegrationsAppleIdentityResolutionProxy) putConversationsMessagingIntegrationsAppleIdentityResolution(ctx context.Context, appleIntegrationId string, config *platformclientv2.Appleidentityresolutionconfig) (appleIdentityResolution *platformclientv2.Appleidentityresolutionconfig, response *platformclientv2.APIResponse, err error) {
	return p.putConversationsMessagingIntegrationsAppleIdentityResolutionAttr(ctx, p, appleIntegrationId, config)
}

func (p *conversationsMessagingIntegrationsAppleIdentityResolutionProxy) getConversationsMessagingIntegrationsAppleById(ctx context.Context, appleIntegrationId string) (appleIntegration *platformclientv2.Appleintegration, response *platformclientv2.APIResponse, err error) {
	return p.getConversationsMessagingIntegrationsAppleByIdAttr(ctx, p, appleIntegrationId)
}

func getConversationsMessagingIntegrationsAppleIdentityResolutionFn(ctx context.Context, p *conversationsMessagingIntegrationsAppleIdentityResolutionProxy, appleIntegrationId string) (appleIdentityResolution *platformclientv2.Appleidentityresolutionconfig, response *platformclientv2.APIResponse, err error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	return p.conversationsApi.GetConversationsMessagingIdentityresolutionIntegrationsAppleIntegrationId(appleIntegrationId)
}

func putConversationsMessagingIntegrationsAppleIdentityResolutionFn(ctx context.Context, p *conversationsMessagingIntegrationsAppleIdentityResolutionProxy, appleIntegrationId string, config *platformclientv2.Appleidentityresolutionconfig) (appleIdentityResolution *platformclientv2.Appleidentityresolutionconfig, response *platformclientv2.APIResponse, err error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	return p.conversationsApi.PutConversationsMessagingIdentityresolutionIntegrationsAppleIntegrationId(appleIntegrationId, *config)
}

func getConversationsMessagingIntegrationsAppleByIdFn(ctx context.Context, p *conversationsMessagingIntegrationsAppleIdentityResolutionProxy, appleIntegrationId string) (appleIntegration *platformclientv2.Appleintegration, response *platformclientv2.APIResponse, err error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	return p.appleIntegrationProxy.GetConversationsMessagingIntegrationsAppleById(ctx, appleIntegrationId)
}
