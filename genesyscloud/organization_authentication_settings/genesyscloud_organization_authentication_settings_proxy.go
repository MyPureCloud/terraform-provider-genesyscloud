package organization_authentication_settings

import (
	"context"
	"fmt"

	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

/*
The genesyscloud_organization_authentication_settings_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *orgAuthSettingsProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getOrgAuthSettingsFunc func(ctx context.Context, p *orgAuthSettingsProxy) (orgAuthSettings *platformclientv2.Orgauthsettings, response *platformclientv2.APIResponse, err error)
type updateOrgAuthSettingsFunc func(ctx context.Context, p *orgAuthSettingsProxy, orgAuthSettings *platformclientv2.Orgauthsettings) (*platformclientv2.Orgauthsettings, *platformclientv2.APIResponse, error)
type getTokensTimeOutSettingsFunc func(ctx context.Context, p *orgAuthSettingsProxy) (*platformclientv2.Idletokentimeout, *platformclientv2.APIResponse, error)
type updateTokensTimeOutSettingsFunc func(ctx context.Context, p *orgAuthSettingsProxy, idletimeout *platformclientv2.Idletokentimeout) (*platformclientv2.Idletokentimeout, *platformclientv2.APIResponse, error)

// orgAuthSettingsProxy contains all of the methods that call genesys cloud APIs.
type orgAuthSettingsProxy struct {
	clientConfig                    *platformclientv2.Configuration
	organizationApi                 *platformclientv2.OrganizationApi
	tokensApi                       *platformclientv2.TokensApi
	getTokensTimeOutSettingsAttr    getTokensTimeOutSettingsFunc
	updateTokensTimeOutSettingsAttr updateTokensTimeOutSettingsFunc
	getOrgAuthSettingsAttr          getOrgAuthSettingsFunc
	updateOrgAuthSettingsAttr       updateOrgAuthSettingsFunc
}

// newOrgAuthSettingsProxy initializes the organization authentication settings proxy with all of the data needed to communicate with Genesys Cloud
func newOrgAuthSettingsProxy(clientConfig *platformclientv2.Configuration) *orgAuthSettingsProxy {
	api := platformclientv2.NewOrganizationApiWithConfig(clientConfig)
	tokenApi := platformclientv2.NewTokensApiWithConfig(clientConfig)
	return &orgAuthSettingsProxy{
		clientConfig:                    clientConfig,
		organizationApi:                 api,
		tokensApi:                       tokenApi,
		getTokensTimeOutSettingsAttr:    getTokensTimeOutSettingsFn,
		updateTokensTimeOutSettingsAttr: updateTokensTimeOutSettingsFn,
		getOrgAuthSettingsAttr:          getOrgAuthSettingsFn,
		updateOrgAuthSettingsAttr:       updateOrgAuthSettingsFn,
	}
}

// getOrgAuthSettingsProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getOrgAuthSettingsProxy(clientConfig *platformclientv2.Configuration) *orgAuthSettingsProxy {
	if internalProxy == nil {
		internalProxy = newOrgAuthSettingsProxy(clientConfig)
	}
	return internalProxy
}

func (p *orgAuthSettingsProxy) getTokensTimeOutSettings(ctx context.Context) (*platformclientv2.Idletokentimeout, *platformclientv2.APIResponse, error) {
	return p.getTokensTimeOutSettingsAttr(ctx, p)
}

func (p *orgAuthSettingsProxy) updateTokensTimeOutSettings(ctx context.Context, idletimeout *platformclientv2.Idletokentimeout) (*platformclientv2.Idletokentimeout, *platformclientv2.APIResponse, error) {
	return p.updateTokensTimeOutSettingsAttr(ctx, p, idletimeout)
}

// getOrgAuthSettings returns a single Genesys Cloud organization authentication settings by Id
func (p *orgAuthSettingsProxy) getOrgAuthSettings(ctx context.Context) (orgAuthSettings *platformclientv2.Orgauthsettings, response *platformclientv2.APIResponse, err error) {
	return p.getOrgAuthSettingsAttr(ctx, p)
}

// updateOrgAuthSettings updates a Genesys Cloud organization authentication settings
func (p *orgAuthSettingsProxy) updateOrgAuthSettings(ctx context.Context, orgAuthSettings *platformclientv2.Orgauthsettings) (*platformclientv2.Orgauthsettings, *platformclientv2.APIResponse, error) {
	return p.updateOrgAuthSettingsAttr(ctx, p, orgAuthSettings)
}

func getTokensTimeOutSettingsFn(ctx context.Context, p *orgAuthSettingsProxy) (*platformclientv2.Idletokentimeout, *platformclientv2.APIResponse, error) {
	idleTokenTimeout, resp, err := p.tokensApi.GetTokensTimeout()
	if err != nil {
		return nil, resp, fmt.Errorf("failed to retrieve token timeout settings: %s", err)
	}
	return idleTokenTimeout, resp, nil
}

func updateTokensTimeOutSettingsFn(ctx context.Context, p *orgAuthSettingsProxy, idletimeout *platformclientv2.Idletokentimeout) (*platformclientv2.Idletokentimeout, *platformclientv2.APIResponse, error) {
	idleTokenTimeout, resp, err := p.tokensApi.PutTokensTimeout(*idletimeout)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to update token timeout settings: %s", err)
	}
	return idleTokenTimeout, resp, nil
}

// getOrgAuthSettingsFn is an implementation of the function to get a Genesys Cloud organization authentication settings by Id
func getOrgAuthSettingsFn(ctx context.Context, p *orgAuthSettingsProxy) (orgAuthSettings *platformclientv2.Orgauthsettings, response *platformclientv2.APIResponse, err error) {
	orgAuthSettings, resp, err := p.organizationApi.GetOrganizationsAuthenticationSettings()
	if err != nil {
		return nil, resp, fmt.Errorf("failed to retrieve organization authentication settings: %s", err)
	}
	return orgAuthSettings, resp, nil
}

// updateOrgAuthSettingsFn is an implementation of the function to update a Genesys Cloud organization authentication settings
func updateOrgAuthSettingsFn(ctx context.Context, p *orgAuthSettingsProxy, orgAuthSettings *platformclientv2.Orgauthsettings) (*platformclientv2.Orgauthsettings, *platformclientv2.APIResponse, error) {
	authSettings, resp, err := p.organizationApi.PatchOrganizationsAuthenticationSettings(*orgAuthSettings)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to update organization authentication settings: %s", err)
	}
	return authSettings, resp, nil
}
