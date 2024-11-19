package organization_authentication_settings

import (
	"context"
	"fmt"

	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
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

// orgAuthSettingsProxy contains all of the methods that call genesys cloud APIs.
type orgAuthSettingsProxy struct {
	clientConfig              *platformclientv2.Configuration
	organizationApi           *platformclientv2.OrganizationApi
	getOrgAuthSettingsAttr    getOrgAuthSettingsFunc
	updateOrgAuthSettingsAttr updateOrgAuthSettingsFunc
}

// newOrgAuthSettingsProxy initializes the organization authentication settings proxy with all of the data needed to communicate with Genesys Cloud
func newOrgAuthSettingsProxy(clientConfig *platformclientv2.Configuration) *orgAuthSettingsProxy {
	api := platformclientv2.NewOrganizationApiWithConfig(clientConfig)
	return &orgAuthSettingsProxy{
		clientConfig:              clientConfig,
		organizationApi:           api,
		getOrgAuthSettingsAttr:    getOrgAuthSettingsFn,
		updateOrgAuthSettingsAttr: updateOrgAuthSettingsFn,
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

// getOrgAuthSettings returns a single Genesys Cloud organization authentication settings by Id
func (p *orgAuthSettingsProxy) getOrgAuthSettings(ctx context.Context) (orgAuthSettings *platformclientv2.Orgauthsettings, response *platformclientv2.APIResponse, err error) {
	return p.getOrgAuthSettingsAttr(ctx, p)
}

// updateOrgAuthSettings updates a Genesys Cloud organization authentication settings
func (p *orgAuthSettingsProxy) updateOrgAuthSettings(ctx context.Context, orgAuthSettings *platformclientv2.Orgauthsettings) (*platformclientv2.Orgauthsettings, *platformclientv2.APIResponse, error) {
	return p.updateOrgAuthSettingsAttr(ctx, p, orgAuthSettings)
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
