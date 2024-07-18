package idp_generic

import (
	"context"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The genesyscloud_idp_generic_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *idpGenericProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getIdpGenericFunc func(ctx context.Context, p *idpGenericProxy) (*platformclientv2.Genericsaml, *platformclientv2.APIResponse, error)
type updateIdpGenericFunc func(ctx context.Context, p *idpGenericProxy, id string, genericSAML *platformclientv2.Genericsaml) (*platformclientv2.Identityprovider, *platformclientv2.APIResponse, error)
type deleteIdpGenericFunc func(ctx context.Context, p *idpGenericProxy, id string) (response *platformclientv2.APIResponse, err error)

// idpGenericProxy contains all of the methods that call genesys cloud APIs.
type idpGenericProxy struct {
	clientConfig         *platformclientv2.Configuration
	identityProviderApi  *platformclientv2.IdentityProviderApi
	getIdpGenericAttr    getIdpGenericFunc
	updateIdpGenericAttr updateIdpGenericFunc
	deleteIdpGenericAttr deleteIdpGenericFunc
}

// newIdpGenericProxy initializes the idp generic proxy with all of the data needed to communicate with Genesys Cloud
func newIdpGenericProxy(clientConfig *platformclientv2.Configuration) *idpGenericProxy {
	api := platformclientv2.NewIdentityProviderApiWithConfig(clientConfig)
	return &idpGenericProxy{
		clientConfig:         clientConfig,
		identityProviderApi:  api,
		getIdpGenericAttr:    getIdpGenericFn,
		updateIdpGenericAttr: updateIdpGenericFn,
		deleteIdpGenericAttr: deleteIdpGenericFn,
	}
}

// getIdpGenericProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getIdpGenericProxy(clientConfig *platformclientv2.Configuration) *idpGenericProxy {
	if internalProxy == nil {
		internalProxy = newIdpGenericProxy(clientConfig)
	}

	return internalProxy
}

// getIdpGeneric retrieves all Genesys Cloud idp generic
func (p *idpGenericProxy) getIdpGeneric(ctx context.Context) (*platformclientv2.Genericsaml, *platformclientv2.APIResponse, error) {
	return p.getIdpGenericAttr(ctx, p)
}

// updateIdpGeneric updates a Genesys Cloud idp generic
func (p *idpGenericProxy) updateIdpGeneric(ctx context.Context, id string, idpGeneric *platformclientv2.Genericsaml) (*platformclientv2.Identityprovider, *platformclientv2.APIResponse, error) {
	return p.updateIdpGenericAttr(ctx, p, id, idpGeneric)
}

// deleteIdpGeneric deletes a Genesys Cloud idp generic by Id
func (p *idpGenericProxy) deleteIdpGeneric(ctx context.Context, id string) (resp *platformclientv2.APIResponse, err error) {
	return p.deleteIdpGenericAttr(ctx, p, id)
}

// getIdpGenericFn is the implementation for retrieving all idp generic in Genesys Cloud
func getIdpGenericFn(ctx context.Context, p *idpGenericProxy) (*platformclientv2.Genericsaml, *platformclientv2.APIResponse, error) {
	return p.identityProviderApi.GetIdentityprovidersGeneric()
}

// updateIdpGenericFn is an implementation of the function to update a Genesys Cloud idp generic
func updateIdpGenericFn(ctx context.Context, p *idpGenericProxy, id string, idpGeneric *platformclientv2.Genericsaml) (*platformclientv2.Identityprovider, *platformclientv2.APIResponse, error) {
	return p.identityProviderApi.PutIdentityprovidersGeneric(*idpGeneric)
}

// deleteIdpGenericFn is an implementation function for deleting a Genesys Cloud idp generic
func deleteIdpGenericFn(ctx context.Context, p *idpGenericProxy, id string) (response *platformclientv2.APIResponse, err error) {
	_, resp, err := p.identityProviderApi.DeleteIdentityprovidersGeneric()
	if err != nil {
		return resp, err
	}

	return resp, err
}
