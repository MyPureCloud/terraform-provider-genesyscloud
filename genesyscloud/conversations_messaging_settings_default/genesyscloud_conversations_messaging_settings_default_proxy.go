package conversations_messaging_settings_default

import (
	"context"

	"github.com/mypurecloud/platform-client-sdk-go/v149/platformclientv2"
)

/*
	The genesyscloud_conversations_messaging_settings_default_proxy.go file contains the proxy structures and methods that interact
	with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
	out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *conversationsMessagingSettingsDefaultProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getConversationsMessagingSettingsDefaultFunc func(ctx context.Context, p *conversationsMessagingSettingsDefaultProxy) (*platformclientv2.Messagingsetting, *platformclientv2.APIResponse, error)
type updateConversationsMessagingSettingsDefaultFunc func(ctx context.Context, p *conversationsMessagingSettingsDefaultProxy, messagingSettingDefaultRequest *platformclientv2.Messagingsettingdefaultrequest) (*platformclientv2.Messagingsetting, *platformclientv2.APIResponse, error)
type deleteConversationsMessagingSettingsDefaultFunc func(ctx context.Context, p *conversationsMessagingSettingsDefaultProxy) (*platformclientv2.APIResponse, error)

// conversationsMessagingSettingsDefaultProxy contains all of the methods that call genesys cloud APIs.
type conversationsMessagingSettingsDefaultProxy struct {
	clientConfig                                    *platformclientv2.Configuration
	conversationsApi                                *platformclientv2.ConversationsApi
	getConversationsMessagingSettingsDefaultAttr    getConversationsMessagingSettingsDefaultFunc
	updateConversationsMessagingSettingsDefaultAttr updateConversationsMessagingSettingsDefaultFunc
	deleteConversationsMessagingSettingsDefaultAttr deleteConversationsMessagingSettingsDefaultFunc
}

// newConversationsMessagingSettingsDefaultProxy initializes the conversations messaging settings default proxy with all of the data needed to communicate with Genesys Cloud
func newConversationsMessagingSettingsDefaultProxy(clientConfig *platformclientv2.Configuration) *conversationsMessagingSettingsDefaultProxy {
	api := platformclientv2.NewConversationsApiWithConfig(clientConfig)
	return &conversationsMessagingSettingsDefaultProxy{
		clientConfig:     clientConfig,
		conversationsApi: api,
		getConversationsMessagingSettingsDefaultAttr:    getConversationsMessagingSettingsDefaultFn,
		updateConversationsMessagingSettingsDefaultAttr: updateConversationsMessagingSettingsDefaultFn,
		deleteConversationsMessagingSettingsDefaultAttr: deleteConversationsMessagingSettingsDefaultFn,
	}
}

// getConversationsMessagingSettingsDefaultProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getConversationsMessagingSettingsDefaultProxy(clientConfig *platformclientv2.Configuration) *conversationsMessagingSettingsDefaultProxy {
	if internalProxy == nil {
		internalProxy = newConversationsMessagingSettingsDefaultProxy(clientConfig)
	}
	return internalProxy
}

// getConversationsMessagingSettingsDefault returns a single Genesys Cloud conversations messaging settings default by Id
func (p *conversationsMessagingSettingsDefaultProxy) getConversationsMessagingSettingsDefault(ctx context.Context) (*platformclientv2.Messagingsetting, *platformclientv2.APIResponse, error) {
	return p.getConversationsMessagingSettingsDefaultAttr(ctx, p)
}

// updateConversationsMessagingSettingsDefault updates a Genesys Cloud conversations messaging settings default
func (p *conversationsMessagingSettingsDefaultProxy) updateConversationsMessagingSettingsDefault(ctx context.Context, conversationsMessagingSettingsDefault *platformclientv2.Messagingsettingdefaultrequest) (*platformclientv2.Messagingsetting, *platformclientv2.APIResponse, error) {
	return p.updateConversationsMessagingSettingsDefaultAttr(ctx, p, conversationsMessagingSettingsDefault)
}

// deleteConversationsMessagingSettingsDefault deletes a Genesys Cloud conversations messaging settings default by Id
func (p *conversationsMessagingSettingsDefaultProxy) deleteConversationsMessagingSettingsDefault(ctx context.Context) (*platformclientv2.APIResponse, error) {
	return p.deleteConversationsMessagingSettingsDefaultAttr(ctx, p)
}

// getConversationsMessagingSettingsDefaultFn is an implementation of the function to get a Genesys Cloud conversations messaging settings default by Id
func getConversationsMessagingSettingsDefaultFn(ctx context.Context, p *conversationsMessagingSettingsDefaultProxy) (*platformclientv2.Messagingsetting, *platformclientv2.APIResponse, error) {
	return p.conversationsApi.GetConversationsMessagingSettingsDefault()
}

// updateConversationsMessagingSettingsDefaultFn is an implementation of the function to update a Genesys Cloud conversations messaging settings default
func updateConversationsMessagingSettingsDefaultFn(ctx context.Context, p *conversationsMessagingSettingsDefaultProxy, conversationsMessagingSettingsDefault *platformclientv2.Messagingsettingdefaultrequest) (*platformclientv2.Messagingsetting, *platformclientv2.APIResponse, error) {
	return p.conversationsApi.PutConversationsMessagingSettingsDefault(*conversationsMessagingSettingsDefault)
}

// deleteConversationsMessagingSettingsDefaultFn is an implementation function for deleting a Genesys Cloud conversations messaging settings default
func deleteConversationsMessagingSettingsDefaultFn(ctx context.Context, p *conversationsMessagingSettingsDefaultProxy) (*platformclientv2.APIResponse, error) {
	return p.conversationsApi.DeleteConversationsMessagingSettingsDefault()
}
