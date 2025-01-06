package conversations_messaging_supportedcontent_default

import (
	"context"

	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
)

/*
The genesyscloud_conversations_messaging_supportedcontent_default_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *conversationsMessagingSupportedcontentDefaultProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getConversationsMessagingSupportedcontentDefaultFunc func(ctx context.Context, p *conversationsMessagingSupportedcontentDefaultProxy) (*platformclientv2.Supportedcontent, *platformclientv2.APIResponse, error)
type updateConversationsMessagingSupportedcontentDefaultFunc func(ctx context.Context, p *conversationsMessagingSupportedcontentDefaultProxy, id string, supportedContentReference *platformclientv2.Supportedcontentreference) (*platformclientv2.Supportedcontent, *platformclientv2.APIResponse, error)

// conversationsMessagingSupportedcontentDefaultProxy contains all of the methods that call genesys cloud APIs.
type conversationsMessagingSupportedcontentDefaultProxy struct {
	clientConfig                                            *platformclientv2.Configuration
	conversationsApi                                        *platformclientv2.ConversationsApi
	getConversationsMessagingSupportedcontentDefaultAttr    getConversationsMessagingSupportedcontentDefaultFunc
	updateConversationsMessagingSupportedcontentDefaultAttr updateConversationsMessagingSupportedcontentDefaultFunc
}

// newConversationsMessagingSupportedcontentDefaultProxy initializes the conversations messaging supportedcontent default proxy with all of the data needed to communicate with Genesys Cloud
func newConversationsMessagingSupportedcontentDefaultProxy(clientConfig *platformclientv2.Configuration) *conversationsMessagingSupportedcontentDefaultProxy {
	api := platformclientv2.NewConversationsApiWithConfig(clientConfig)
	return &conversationsMessagingSupportedcontentDefaultProxy{
		clientConfig:     clientConfig,
		conversationsApi: api,
		getConversationsMessagingSupportedcontentDefaultAttr:    getConversationsMessagingSupportedcontentDefaultFn,
		updateConversationsMessagingSupportedcontentDefaultAttr: updateConversationsMessagingSupportedcontentDefaultFn,
	}
}

// getConversationsMessagingSupportedcontentDefaultProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getConversationsMessagingSupportedcontentDefaultProxy(clientConfig *platformclientv2.Configuration) *conversationsMessagingSupportedcontentDefaultProxy {
	if internalProxy == nil {
		internalProxy = newConversationsMessagingSupportedcontentDefaultProxy(clientConfig)
	}

	return internalProxy
}

// getConversationsMessagingSupportedcontentDefault retrieves all Genesys Cloud conversations messaging supportedcontent default
func (p *conversationsMessagingSupportedcontentDefaultProxy) getConversationsMessagingSupportedcontentDefault(ctx context.Context) (*platformclientv2.Supportedcontent, *platformclientv2.APIResponse, error) {
	return p.getConversationsMessagingSupportedcontentDefaultAttr(ctx, p)
}

// updateConversationsMessagingSupportedcontentDefault updates a Genesys Cloud conversations messaging supportedcontent default
func (p *conversationsMessagingSupportedcontentDefaultProxy) updateConversationsMessagingSupportedcontentDefault(ctx context.Context, id string, conversationsMessagingSupportedcontentDefault *platformclientv2.Supportedcontentreference) (*platformclientv2.Supportedcontent, *platformclientv2.APIResponse, error) {
	return p.updateConversationsMessagingSupportedcontentDefaultAttr(ctx, p, id, conversationsMessagingSupportedcontentDefault)
}

// getConversationsMessagingSupportedcontentDefaultFn is the implementation for retrieving all conversations messaging supportedcontent default in Genesys Cloud
func getConversationsMessagingSupportedcontentDefaultFn(ctx context.Context, p *conversationsMessagingSupportedcontentDefaultProxy) (*platformclientv2.Supportedcontent, *platformclientv2.APIResponse, error) {
	return p.conversationsApi.GetConversationsMessagingSupportedcontentDefault()
}

// updateConversationsMessagingSupportedcontentDefaultFn is an implementation of the function to update a Genesys Cloud conversations messaging supportedcontent default
func updateConversationsMessagingSupportedcontentDefaultFn(ctx context.Context, p *conversationsMessagingSupportedcontentDefaultProxy, id string, conversationsMessagingSupportedcontentDefault *platformclientv2.Supportedcontentreference) (*platformclientv2.Supportedcontent, *platformclientv2.APIResponse, error) {
	return p.conversationsApi.PutConversationsMessagingSupportedcontentDefault(*conversationsMessagingSupportedcontentDefault)
}
