package organization_authentication_settings

import (
	"context"
	"fmt"
	"github.com/mypurecloud/platform-client-sdk-go/v121/platformclientv2"
)

/*
The genesyscloud_organization_authentication_settings_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *orgAuthSettingsProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getOrgAuthSettingsByIdFunc func(ctx context.Context, p *orgAuthSettingsProxy, id string) (orgAuthSettings *platformclientv2.Orgauthsettings, responseCode int, err error)
type updateOrgAuthSettingsFunc func(ctx context.Context, p *orgAuthSettingsProxy, orgAuthSettings *platformclientv2.Orgauthsettings) (*platformclientv2.Orgauthsettings, int, error)

// orgAuthSettingsProxy contains all of the methods that call genesys cloud APIs.
type orgAuthSettingsProxy struct {
	clientConfig               *platformclientv2.Configuration
	organizationApi            *platformclientv2.OrganizationApi
	getOrgAuthSettingsByIdAttr getOrgAuthSettingsByIdFunc
	updateOrgAuthSettingsAttr  updateOrgAuthSettingsFunc
}

// newOrgAuthSettingsProxy initializes the organization authentication settings proxy with all of the data needed to communicate with Genesys Cloud
func newOrgAuthSettingsProxy(clientConfig *platformclientv2.Configuration) *orgAuthSettingsProxy {
	api := platformclientv2.NewOrganizationApiWithConfig(clientConfig)
	return &orgAuthSettingsProxy{
		clientConfig:               clientConfig,
		organizationApi:            api,
		getOrgAuthSettingsByIdAttr: getOrgAuthSettingsByIdFn,
		updateOrgAuthSettingsAttr:  updateOrgAuthSettingsFn,
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

// getOrgAuthSettingsById returns a single Genesys Cloud organization authentication settings by Id
func (p *orgAuthSettingsProxy) getOrgAuthSettingsById(ctx context.Context, id string) (orgAuthSettings *platformclientv2.Orgauthsettings, statusCode int, err error) {
	return p.getOrgAuthSettingsByIdAttr(ctx, p, id)
}

// updateOrgAuthSettings updates a Genesys Cloud organization authentication settings
func (p *orgAuthSettingsProxy) updateOrgAuthSettings(ctx context.Context, orgAuthSettings *platformclientv2.Orgauthsettings) (*platformclientv2.Orgauthsettings, int, error) {
	return p.updateOrgAuthSettingsAttr(ctx, p, orgAuthSettings)
}

// getOrgAuthSettingsByIdFn is an implementation of the function to get a Genesys Cloud organization authentication settings by Id
func getOrgAuthSettingsByIdFn(ctx context.Context, p *orgAuthSettingsProxy, id string) (orgAuthSettings *platformclientv2.Orgauthsettings, statusCode int, err error) {
	orgAuthSettings, resp, err := p.organizationApi.GetOrganizationsAuthenticationSettings()
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("Failed to retrieve organization authentication settings by id %s: %s", id, err)
	}
	return orgAuthSettings, resp.StatusCode, nil
}

// updateOrgAuthSettingsFn is an implementation of the function to update a Genesys Cloud organization authentication settings
func updateOrgAuthSettingsFn(ctx context.Context, p *orgAuthSettingsProxy, orgAuthSettings *platformclientv2.Orgauthsettings) (*platformclientv2.Orgauthsettings, int, error) {
	authSettings, resp, err := p.organizationApi.PatchOrganizationsAuthenticationSettings(*orgAuthSettings)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to update organization authentication settings: %s", err)
	}
	return authSettings, resp.StatusCode, nil
}
