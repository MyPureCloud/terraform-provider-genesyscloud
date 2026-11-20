package authorization_roles_settings

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"

	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
)

/*
The genesyscloud_authorization_roles_settings_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// rolesSettingsPath is the API path backing this resource.
const rolesSettingsPath = "/api/v2/authorization/roles/settings"

// roleSettings mirrors the RoleSettings model of /api/v2/authorization/roles/settings.
//
// The generated platformclientv2.Rolesettings struct is deliberately NOT used here. The v200
// SDK this provider pins predates genesysOrgPolicyBypass and so has no field for it, which
// would make the attribute this resource exists to manage unreachable. Declaring the model
// locally keeps the whole change contained to this package instead of requiring a repo-wide
// bump to the v201 module path.
//
// The requests themselves still go through the provider's own authenticated
// platformclientv2.Configuration, so pooling, proxy and mTLS settings, retries and the 401
// token refresh all behave exactly as they do for generated SDK calls.
//
// When the provider's SDK dependency is eventually bumped to v201 or later, this struct and
// the two call helpers below can be deleted in favour of the generated
// GetAuthorizationRolesSettings / PutAuthorizationRolesSettings methods without touching any
// other layer of the package.
type roleSettings struct {
	Id *string `json:"id,omitempty"`
	// BackfillEnabled - Boolean showing if organization is opted in or not to role backfills
	BackfillEnabled *bool `json:"backfillEnabled,omitempty"`
	// AuthorizationGrantDivisionAware - Boolean enabling enforcement of division aware for authorization grant add and delete
	AuthorizationGrantDivisionAware *bool `json:"authorizationGrantDivisionAware,omitempty"`
	// GenesysOrgPolicyBypass - Boolean enabling skip of attribute-based access control policy enforcement when enabled and the organization is on the Genesys bypass list
	GenesysOrgPolicyBypass *bool   `json:"genesysOrgPolicyBypass,omitempty"`
	SelfUri                *string `json:"selfUri,omitempty"`
}

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *authorizationRolesSettingsProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getAuthorizationRolesSettingsFunc func(ctx context.Context, p *authorizationRolesSettingsProxy) (*roleSettings, *platformclientv2.APIResponse, error)
type updateAuthorizationRolesSettingsFunc func(ctx context.Context, p *authorizationRolesSettingsProxy, settings *roleSettings) (*roleSettings, *platformclientv2.APIResponse, error)

// authorizationRolesSettingsProxy contains all of the methods that call genesys cloud APIs.
type authorizationRolesSettingsProxy struct {
	clientConfig                         *platformclientv2.Configuration
	getAuthorizationRolesSettingsAttr    getAuthorizationRolesSettingsFunc
	updateAuthorizationRolesSettingsAttr updateAuthorizationRolesSettingsFunc
}

// newAuthorizationRolesSettingsProxy initializes the proxy with all of the data needed to communicate with Genesys Cloud
func newAuthorizationRolesSettingsProxy(clientConfig *platformclientv2.Configuration) *authorizationRolesSettingsProxy {
	return &authorizationRolesSettingsProxy{
		clientConfig:                         clientConfig,
		getAuthorizationRolesSettingsAttr:    getAuthorizationRolesSettingsFn,
		updateAuthorizationRolesSettingsAttr: updateAuthorizationRolesSettingsFn,
	}
}

// getAuthorizationRolesSettingsProxy acts as a singleton for the internalProxy. It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getAuthorizationRolesSettingsProxy(clientConfig *platformclientv2.Configuration) *authorizationRolesSettingsProxy {
	return provider.SingletonOrFresh(&internalProxy, clientConfig, newAuthorizationRolesSettingsProxy)
}

// getAuthorizationRolesSettings returns the Genesys Cloud organization authorization role settings
func (p *authorizationRolesSettingsProxy) getAuthorizationRolesSettings(ctx context.Context) (*roleSettings, *platformclientv2.APIResponse, error) {
	return p.getAuthorizationRolesSettingsAttr(ctx, p)
}

// updateAuthorizationRolesSettings updates the Genesys Cloud organization authorization role settings
func (p *authorizationRolesSettingsProxy) updateAuthorizationRolesSettings(ctx context.Context, settings *roleSettings) (*roleSettings, *platformclientv2.APIResponse, error) {
	return p.updateAuthorizationRolesSettingsAttr(ctx, p, settings)
}

// getAuthorizationRolesSettingsFn is an implementation of the function to get the Genesys Cloud authorization role settings
func getAuthorizationRolesSettingsFn(ctx context.Context, p *authorizationRolesSettingsProxy) (*roleSettings, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	settings, resp, err := p.callRoleSettingsAPI("GET", nil)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to retrieve authorization role settings: %s", err)
	}
	return settings, resp, nil
}

// updateAuthorizationRolesSettingsFn is an implementation of the function to update the Genesys Cloud authorization role settings.
//
// PUT replaces the whole settings object, so callers must supply a fully populated one. The
// resource layer builds that by merging configuration over a fresh read; see
// buildAuthorizationRolesSettingsFromResourceData.
func updateAuthorizationRolesSettingsFn(ctx context.Context, p *authorizationRolesSettingsProxy, settings *roleSettings) (*roleSettings, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	updated, resp, err := p.callRoleSettingsAPI("PUT", settings)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to update authorization role settings: %s", err)
	}
	return updated, resp, nil
}

// callRoleSettingsAPI issues a request against the role settings endpoint using the provider's
// configured SDK client.
//
// It mirrors what the generated platformclientv2 AuthorizationApi methods do: resolve the path
// from the configuration's BasePath, attach the bearer token and any default headers, negotiate
// content types through the shared APIClient, and hand back the standard *APIResponse so
// callers can keep using util.IsStatus404 and util.BuildAPIDiagnosticError. Because the call
// goes through Configuration.APIClient, retry, proxy, gateway, mTLS and the 401-triggered token
// refresh are all handled by the SDK rather than reimplemented here.
func (p *authorizationRolesSettingsProxy) callRoleSettingsAPI(method string, body *roleSettings) (*roleSettings, *platformclientv2.APIResponse, error) {
	if p.clientConfig == nil {
		return nil, nil, errors.New("no SDK client configuration available")
	}

	apiClient := &p.clientConfig.APIClient
	path := p.clientConfig.BasePath + rolesSettingsPath

	headerParams := make(map[string]string)
	if p.clientConfig.AccessToken != "" {
		headerParams["Authorization"] = "Bearer " + p.clientConfig.AccessToken
	}
	for key := range p.clientConfig.DefaultHeader {
		headerParams[key] = p.clientConfig.DefaultHeader[key]
	}
	if contentType := apiClient.SelectHeaderContentType([]string{"application/json"}); contentType != "" {
		headerParams["Content-Type"] = contentType
	}
	if accept := apiClient.SelectHeaderAccept([]string{"application/json"}); accept != "" {
		headerParams["Accept"] = accept
	}

	var postBody interface{}
	if body != nil {
		postBody = body
	}

	response, err := apiClient.CallAPI(path, method, postBody, headerParams, make(map[string]string), url.Values{}, "", nil, "other")
	if err != nil {
		return nil, response, err
	}
	if response.Error != nil {
		return nil, response, errors.New(response.ErrorMessage)
	}

	var successPayload *roleSettings
	if response.HasBody {
		if err := json.Unmarshal(response.RawBody, &successPayload); err != nil {
			return nil, response, err
		}
	}
	return successPayload, response, nil
}
