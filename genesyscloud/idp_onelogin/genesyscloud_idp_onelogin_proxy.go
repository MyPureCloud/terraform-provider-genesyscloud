package idp_onelogin

import (
	"context"
	"fmt"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The genesyscloud_idp_onelogin_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *idpOneloginProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getIdpOneloginFunc func(ctx context.Context, p *idpOneloginProxy) (*platformclientv2.Onelogin, *platformclientv2.APIResponse, error)
type updateIdpOneloginFunc func(ctx context.Context, p *idpOneloginProxy, id string, oneLogin *platformclientv2.Onelogin) (*platformclientv2.Identityprovider, *platformclientv2.APIResponse, error)
type deleteIdpOneloginFunc func(ctx context.Context, p *idpOneloginProxy, id string) (response *platformclientv2.APIResponse, err error)

// idpOneloginProxy contains all of the methods that call genesys cloud APIs.
type idpOneloginProxy struct {
	clientConfig          *platformclientv2.Configuration
	identityProviderApi   *platformclientv2.IdentityProviderApi
	getIdpOneloginAttr    getIdpOneloginFunc
	updateIdpOneloginAttr updateIdpOneloginFunc
	deleteIdpOneloginAttr deleteIdpOneloginFunc
}

// newIdpOneloginProxy initializes the idp onelogin proxy with all of the data needed to communicate with Genesys Cloud
func newIdpOneloginProxy(clientConfig *platformclientv2.Configuration) *idpOneloginProxy {
	api := platformclientv2.NewIdentityProviderApiWithConfig(clientConfig)
	return &idpOneloginProxy{
		clientConfig:          clientConfig,
		identityProviderApi:   api,
		getIdpOneloginAttr:    getIdpOneloginFn,
		updateIdpOneloginAttr: updateIdpOneloginFn,
		deleteIdpOneloginAttr: deleteIdpOneloginFn,
	}
}

// getIdpOneloginProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getIdpOneloginProxy(clientConfig *platformclientv2.Configuration) *idpOneloginProxy {
	if internalProxy == nil {
		internalProxy = newIdpOneloginProxy(clientConfig)
	}

	return internalProxy
}

// getIdpOnelogin retrieves all Genesys Cloud idp onelogin
func (p *idpOneloginProxy) getIdpOnelogin(ctx context.Context) (*platformclientv2.Onelogin, *platformclientv2.APIResponse, error) {
	return p.getIdpOneloginAttr(ctx, p)
}

// updateIdpOnelogin updates a Genesys Cloud idp onelogin
func (p *idpOneloginProxy) updateIdpOnelogin(ctx context.Context, id string, idpOnelogin *platformclientv2.Onelogin) (*platformclientv2.Identityprovider, *platformclientv2.APIResponse, error) {
	return p.updateIdpOneloginAttr(ctx, p, id, idpOnelogin)
}

// deleteIdpOnelogin deletes a Genesys Cloud idp onelogin by Id
func (p *idpOneloginProxy) deleteIdpOnelogin(ctx context.Context, id string) (response *platformclientv2.APIResponse, err error) {
	return p.deleteIdpOneloginAttr(ctx, p, id)
}

// getAllIdpOneloginFn is the implementation for retrieving all idp onelogin in Genesys Cloud
func getIdpOneloginFn(ctx context.Context, p *idpOneloginProxy) (*platformclientv2.Onelogin, *platformclientv2.APIResponse, error) {
	return p.identityProviderApi.GetIdentityprovidersOnelogin()
}

// updateIdpOneloginFn is an implementation of the function to update a Genesys Cloud idp onelogin
func updateIdpOneloginFn(ctx context.Context, p *idpOneloginProxy, id string, idpOnelogin *platformclientv2.Onelogin) (*platformclientv2.Identityprovider, *platformclientv2.APIResponse, error) {
	return p.identityProviderApi.PutIdentityprovidersOnelogin(*idpOnelogin)
}

// deleteIdpOneloginFn is an implementation function for deleting a Genesys Cloud idp onelogin
func deleteIdpOneloginFn(ctx context.Context, p *idpOneloginProxy, id string) (response *platformclientv2.APIResponse, err error) {
	_, resp, err := p.identityProviderApi.DeleteIdentityprovidersOnelogin()
	if err != nil {
		return resp, fmt.Errorf("Failed to delete idp onelogin: %s", err)
	}

	return resp, err
}
