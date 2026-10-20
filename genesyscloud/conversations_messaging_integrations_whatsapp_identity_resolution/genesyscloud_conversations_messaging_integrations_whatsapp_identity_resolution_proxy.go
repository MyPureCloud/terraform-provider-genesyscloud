package conversations_messaging_integrations_whatsapp_identity_resolution

import (
	"context"

	"github.com/mypurecloud/platform-client-sdk-go/v199/platformclientv2"
	whatsapp "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_integrations_whatsapp"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
)

/*
The genesyscloud_conversations_messaging_integrations_whatsapp_identity_resolution_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

var internalProxy *conversationsMessagingIntegrationsWhatsappIdentityResolutionProxy

type getConversationsMessagingIntegrationsWhatsappIdentityResolutionFunc func(ctx context.Context, p *conversationsMessagingIntegrationsWhatsappIdentityResolutionProxy, whatsappIntegrationId string) (whatsappIdentityResolution *platformclientv2.Whatsappidentityresolutionconfig, response *platformclientv2.APIResponse, err error)
type putConversationsMessagingIntegrationsWhatsappIdentityResolutionFunc func(ctx context.Context, p *conversationsMessagingIntegrationsWhatsappIdentityResolutionProxy, whatsappIntegrationId string, config *platformclientv2.Whatsappidentityresolutionconfig) (whatsappIdentityResolution *platformclientv2.Whatsappidentityresolutionconfig, response *platformclientv2.APIResponse, err error)
type getConversationsMessagingIntegrationsWhatsappByIdFunc func(ctx context.Context, p *conversationsMessagingIntegrationsWhatsappIdentityResolutionProxy, whatsappIntegrationId string) (whatsappIntegration *platformclientv2.Whatsappintegration, response *platformclientv2.APIResponse, err error)

type conversationsMessagingIntegrationsWhatsappIdentityResolutionProxy struct {
	clientConfig                                                        *platformclientv2.Configuration
	conversationsApi                                                    *platformclientv2.ConversationsApi
	getConversationsMessagingIntegrationsWhatsappIdentityResolutionAttr getConversationsMessagingIntegrationsWhatsappIdentityResolutionFunc
	putConversationsMessagingIntegrationsWhatsappIdentityResolutionAttr putConversationsMessagingIntegrationsWhatsappIdentityResolutionFunc
	getConversationsMessagingIntegrationsWhatsappByIdAttr               getConversationsMessagingIntegrationsWhatsappByIdFunc
	whatsappIntegrationProxy                                            *whatsapp.ConversationsMessagingIntegrationsWhatsappProxy
}

func newConversationsMessagingIntegrationsWhatsappIdentityResolutionProxy(clientConfig *platformclientv2.Configuration) *conversationsMessagingIntegrationsWhatsappIdentityResolutionProxy {
	api := platformclientv2.NewConversationsApiWithConfig(clientConfig)
	whatsappIntegrationProxt := whatsapp.GetConversationsMessagingIntegrationsWhatsappProxy(clientConfig)

	return &conversationsMessagingIntegrationsWhatsappIdentityResolutionProxy{
		clientConfig:     clientConfig,
		conversationsApi: api,
		getConversationsMessagingIntegrationsWhatsappIdentityResolutionAttr: getConversationsMessagingIntegrationsWhatsappIdentityResolutionFn,
		putConversationsMessagingIntegrationsWhatsappIdentityResolutionAttr: putConversationsMessagingIntegrationsWhatsappIdentityResolutionFn,
		getConversationsMessagingIntegrationsWhatsappByIdAttr:               getConversationsMessagingIntegrationsWhatsappByIdFn,
		whatsappIntegrationProxy: whatsappIntegrationProxt,
	}
}

func getConversationsMessagingIntegrationsWhatsappIdentityResolutionProxy(clientConfig *platformclientv2.Configuration) *conversationsMessagingIntegrationsWhatsappIdentityResolutionProxy {
	if internalProxy == nil {
		internalProxy = newConversationsMessagingIntegrationsWhatsappIdentityResolutionProxy(clientConfig)
	}
	return internalProxy
}

func (p *conversationsMessagingIntegrationsWhatsappIdentityResolutionProxy) getConversationsMessagingIntegrationsWhatsappIdentityResolution(ctx context.Context, whatsappIntegrationId string) (whatsappIdentityResolution *platformclientv2.Whatsappidentityresolutionconfig, response *platformclientv2.APIResponse, err error) {
	return p.getConversationsMessagingIntegrationsWhatsappIdentityResolutionAttr(ctx, p, whatsappIntegrationId)
}

func (p *conversationsMessagingIntegrationsWhatsappIdentityResolutionProxy) putConversationsMessagingIntegrationsWhatsappIdentityResolution(ctx context.Context, whatsappIntegrationId string, config *platformclientv2.Whatsappidentityresolutionconfig) (whatsappIdentityResolution *platformclientv2.Whatsappidentityresolutionconfig, response *platformclientv2.APIResponse, err error) {
	return p.putConversationsMessagingIntegrationsWhatsappIdentityResolutionAttr(ctx, p, whatsappIntegrationId, config)
}

func (p *conversationsMessagingIntegrationsWhatsappIdentityResolutionProxy) getConversationsMessagingIntegrationsWhatsappById(ctx context.Context, whatsappIntegrationId string) (whatsappIntegration *platformclientv2.Whatsappintegration, response *platformclientv2.APIResponse, err error) {
	return p.getConversationsMessagingIntegrationsWhatsappByIdAttr(ctx, p, whatsappIntegrationId)
}

func getConversationsMessagingIntegrationsWhatsappIdentityResolutionFn(ctx context.Context, p *conversationsMessagingIntegrationsWhatsappIdentityResolutionProxy, whatsappIntegrationId string) (whatsappIdentityResolution *platformclientv2.Whatsappidentityresolutionconfig, response *platformclientv2.APIResponse, err error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	return p.conversationsApi.GetConversationsMessagingIdentityresolutionIntegrationsWhatsappIntegrationId(whatsappIntegrationId)
}

func putConversationsMessagingIntegrationsWhatsappIdentityResolutionFn(ctx context.Context, p *conversationsMessagingIntegrationsWhatsappIdentityResolutionProxy, whatsappIntegrationId string, config *platformclientv2.Whatsappidentityresolutionconfig) (whatsappIdentityResolution *platformclientv2.Whatsappidentityresolutionconfig, response *platformclientv2.APIResponse, err error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	return p.conversationsApi.PutConversationsMessagingIdentityresolutionIntegrationsWhatsappIntegrationId(whatsappIntegrationId, *config)
}

func getConversationsMessagingIntegrationsWhatsappByIdFn(ctx context.Context, p *conversationsMessagingIntegrationsWhatsappIdentityResolutionProxy, whatsappIntegrationId string) (whatsappIntegration *platformclientv2.Whatsappintegration, response *platformclientv2.APIResponse, err error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	return p.whatsappIntegrationProxy.GetConversationsMessagingIntegrationsWhatsappById(ctx, whatsappIntegrationId)
}
