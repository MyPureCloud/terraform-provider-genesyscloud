package conversations_settings

import (
	"context"

	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"
)

var internalProxy *conversationsSettingsProxy

type getConversationsSettingsFunc func(ctx context.Context, p *conversationsSettingsProxy) (*platformclientv2.Settings, *platformclientv2.APIResponse, error)
type updateConversationsSettingsFunc func(ctx context.Context, p *conversationsSettingsProxy, settings *platformclientv2.Settings) (*platformclientv2.APIResponse, error)

// conversationsSettingsProxy contains all of the methods that call genesys cloud APIs.
type conversationsSettingsProxy struct {
	clientConfig                    *platformclientv2.Configuration
	conversationsApi                *platformclientv2.ConversationsApi
	getConversationsSettingsAttr    getConversationsSettingsFunc
	updateConversationsSettingsAttr updateConversationsSettingsFunc
}

// newConversationsSettingsProxy initializes the conversations settings proxy with all of the data needed to communicate with Genesys Cloud
func newConversationsSettingsProxy(clientConfig *platformclientv2.Configuration) *conversationsSettingsProxy {
	api := platformclientv2.NewConversationsApiWithConfig(clientConfig)
	return &conversationsSettingsProxy{
		clientConfig:                    clientConfig,
		conversationsApi:                api,
		getConversationsSettingsAttr:    getConversationsSettingsFn,
		updateConversationsSettingsAttr: updateConversationsSettingsFn,
	}
}

// getConversationsSettingsProxy acts as a singleton to for the internalProxy. It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getConversationsSettingsProxy(clientConfig *platformclientv2.Configuration) *conversationsSettingsProxy {
	if internalProxy == nil {
		internalProxy = newConversationsSettingsProxy(clientConfig)
	}
	return internalProxy
}

// getConversationsSettings retrieves the Genesys Cloud conversations settings
func (p *conversationsSettingsProxy) getConversationsSettings(ctx context.Context) (*platformclientv2.Settings, *platformclientv2.APIResponse, error) {
	return p.getConversationsSettingsAttr(ctx, p)
}

// updateConversationsSettings updates the Genesys Cloud conversations settings
func (p *conversationsSettingsProxy) updateConversationsSettings(ctx context.Context, settings *platformclientv2.Settings) (*platformclientv2.APIResponse, error) {
	return p.updateConversationsSettingsAttr(ctx, p, settings)
}

// getConversationsSettingsFn is the implementation for retrieving conversations settings from Genesys Cloud
func getConversationsSettingsFn(ctx context.Context, p *conversationsSettingsProxy) (*platformclientv2.Settings, *platformclientv2.APIResponse, error) {
	return p.conversationsApi.GetConversationsSettings()
}

// updateConversationsSettingsFn is the implementation for updating conversations settings in Genesys Cloud
func updateConversationsSettingsFn(ctx context.Context, p *conversationsSettingsProxy, settings *platformclientv2.Settings) (*platformclientv2.APIResponse, error) {
	return p.conversationsApi.PatchConversationsSettings(*settings)
}
