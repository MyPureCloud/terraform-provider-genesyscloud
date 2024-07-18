package idp_okta

import (
	"context"
	"fmt"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The genesyscloud_idp_okta_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *idpOktaProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getIdpOktaFunc func(ctx context.Context, p *idpOktaProxy) (*platformclientv2.Okta, *platformclientv2.APIResponse, error)
type updateIdpOktaFunc func(ctx context.Context, p *idpOktaProxy, id string, okta *platformclientv2.Okta) (*platformclientv2.Identityprovider, *platformclientv2.APIResponse, error)
type deleteIdpOktaFunc func(ctx context.Context, p *idpOktaProxy, id string) (response *platformclientv2.APIResponse, err error)

// idpOktaProxy contains all of the methods that call genesys cloud APIs.
type idpOktaProxy struct {
	clientConfig        *platformclientv2.Configuration
	identityProviderApi *platformclientv2.IdentityProviderApi
	getIdpOktaAttr      getIdpOktaFunc
	updateIdpOktaAttr   updateIdpOktaFunc
	deleteIdpOktaAttr   deleteIdpOktaFunc
}

// newIdpOktaProxy initializes the idp okta proxy with all of the data needed to communicate with Genesys Cloud
func newIdpOktaProxy(clientConfig *platformclientv2.Configuration) *idpOktaProxy {
	api := platformclientv2.NewIdentityProviderApiWithConfig(clientConfig)
	return &idpOktaProxy{
		clientConfig:        clientConfig,
		identityProviderApi: api,
		getIdpOktaAttr:      getIdpOktaFn,
		updateIdpOktaAttr:   updateIdpOktaFn,
		deleteIdpOktaAttr:   deleteIdpOktaFn,
	}
}

// getIdpOktaProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getIdpOktaProxy(clientConfig *platformclientv2.Configuration) *idpOktaProxy {
	if internalProxy == nil {
		internalProxy = newIdpOktaProxy(clientConfig)
	}

	return internalProxy
}

// getIdpOkta retrieves all Genesys Cloud idp okta
func (p *idpOktaProxy) getIdpOkta(ctx context.Context) (*platformclientv2.Okta, *platformclientv2.APIResponse, error) {
	return p.getIdpOktaAttr(ctx, p)
}

// updateIdpOkta updates a Genesys Cloud idp okta
func (p *idpOktaProxy) updateIdpOkta(ctx context.Context, id string, idpOkta *platformclientv2.Okta) (*platformclientv2.Identityprovider, *platformclientv2.APIResponse, error) {
	return p.updateIdpOktaAttr(ctx, p, id, idpOkta)
}

// deleteIdpOkta deletes a Genesys Cloud idp okta by Id
func (p *idpOktaProxy) deleteIdpOkta(ctx context.Context, id string) (response *platformclientv2.APIResponse, err error) {
	return p.deleteIdpOktaAttr(ctx, p, id)
}

// getIdpOktaFn is the implementation for retrieving all idp okta in Genesys Cloud
func getIdpOktaFn(ctx context.Context, p *idpOktaProxy) (*platformclientv2.Okta, *platformclientv2.APIResponse, error) {
	return p.identityProviderApi.GetIdentityprovidersOkta()
}

// updateIdpOktaFn is an implementation of the function to update a Genesys Cloud idp okta
func updateIdpOktaFn(ctx context.Context, p *idpOktaProxy, id string, idpOkta *platformclientv2.Okta) (*platformclientv2.Identityprovider, *platformclientv2.APIResponse, error) {
	return p.identityProviderApi.PutIdentityprovidersOkta(*idpOkta)
}

// deleteIdpOktaFn is an implementation function for deleting a Genesys Cloud idp okta
func deleteIdpOktaFn(ctx context.Context, p *idpOktaProxy, id string) (response *platformclientv2.APIResponse, err error) {
	_, resp, err := p.identityProviderApi.DeleteIdentityprovidersOkta()
	if err != nil {
		return resp, fmt.Errorf("Failed to delete idp okta: %s", err)
	}

	return resp, nil
}
