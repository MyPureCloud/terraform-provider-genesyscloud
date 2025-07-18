package outbound_settings

import (
	"context"
	"fmt"

	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"
)

/*
The genesyscloud_outbound_settings_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *outboundSettingsProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getOutboundSettingsFunc func(ctx context.Context, p *outboundSettingsProxy) (*platformclientv2.Outboundsettings, *platformclientv2.APIResponse, error)
type updateOutboundSettingsFunc func(ctx context.Context, p *outboundSettingsProxy, outboundSettings *platformclientv2.Outboundsettings) (*platformclientv2.Outboundsettings, *platformclientv2.APIResponse, error)

// outboundSettingsProxy contains all of the methods that call genesys cloud APIs.
type outboundSettingsProxy struct {
	clientConfig               *platformclientv2.Configuration
	outboundApi                *platformclientv2.OutboundApi
	getOutboundSettingsAttr    getOutboundSettingsFunc
	updateOutboundSettingsAttr updateOutboundSettingsFunc
}

// newOutboundSettingsProxy initializes the outbound settings proxy with all of the data needed to communicate with Genesys Cloud
func newOutboundSettingsProxy(clientConfig *platformclientv2.Configuration) *outboundSettingsProxy {
	api := platformclientv2.NewOutboundApiWithConfig(clientConfig)
	return &outboundSettingsProxy{
		clientConfig:               clientConfig,
		outboundApi:                api,
		getOutboundSettingsAttr:    getOutboundSettingsFn,
		updateOutboundSettingsAttr: updateOutboundSettingsFn,
	}
}

// getOutboundSettingsProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getOutboundSettingsProxy(clientConfig *platformclientv2.Configuration) *outboundSettingsProxy {
	if internalProxy == nil {
		internalProxy = newOutboundSettingsProxy(clientConfig)
	}
	return internalProxy
}

// getOutboundSettings returns a single Genesys Cloud outbound settings by Id
func (p *outboundSettingsProxy) getOutboundSettings(ctx context.Context) (*platformclientv2.Outboundsettings, *platformclientv2.APIResponse, error) {
	return p.getOutboundSettingsAttr(ctx, p)
}

// updateOutboundSettings updates a Genesys Cloud outbound settings
func (p *outboundSettingsProxy) updateOutboundSettings(ctx context.Context, outboundSettings *platformclientv2.Outboundsettings) (*platformclientv2.Outboundsettings, *platformclientv2.APIResponse, error) {
	return p.updateOutboundSettingsAttr(ctx, p, outboundSettings)
}

// getOutboundSettingsFn is an implementation of the function to get a Genesys Cloud outbound settings by Id
func getOutboundSettingsFn(ctx context.Context, p *outboundSettingsProxy) (*platformclientv2.Outboundsettings, *platformclientv2.APIResponse, error) {
	outboundSettings, resp, err := p.outboundApi.GetOutboundSettings()
	if err != nil {
		return nil, resp, fmt.Errorf("failed to retrieve outbound settings: %s", err)
	}
	return outboundSettings, resp, nil
}

// updateOutboundSettingsFn is an implementation of the function to update a Genesys Cloud outbound settings
func updateOutboundSettingsFn(_ context.Context, p *outboundSettingsProxy, outboundSettings *platformclientv2.Outboundsettings) (*platformclientv2.Outboundsettings, *platformclientv2.APIResponse, error) {
	resp, err := p.outboundApi.PatchOutboundSettings(*outboundSettings, false)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update outbound settings: %s", err)
	}
	return outboundSettings, resp, nil
}
