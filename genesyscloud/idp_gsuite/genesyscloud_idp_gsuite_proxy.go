package idp_gsuite

import (
	"context"
	"fmt"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The genesyscloud_idp_gsuite_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *idpGsuiteProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getIdpGsuiteFunc func(ctx context.Context, p *idpGsuiteProxy) (*platformclientv2.Gsuite, *platformclientv2.APIResponse, error)
type updateIdpGsuiteFunc func(ctx context.Context, p *idpGsuiteProxy, id string, gSuite *platformclientv2.Gsuite) (*platformclientv2.Identityprovider, *platformclientv2.APIResponse, error)
type deleteIdpGsuiteFunc func(ctx context.Context, p *idpGsuiteProxy, id string) (response *platformclientv2.APIResponse, err error)

// idpGsuiteProxy contains all of the methods that call genesys cloud APIs.
type idpGsuiteProxy struct {
	clientConfig        *platformclientv2.Configuration
	identityProviderApi *platformclientv2.IdentityProviderApi
	getIdpGsuiteAttr    getIdpGsuiteFunc
	updateIdpGsuiteAttr updateIdpGsuiteFunc
	deleteIdpGsuiteAttr deleteIdpGsuiteFunc
}

// newIdpGsuiteProxy initializes the idp gsuite proxy with all of the data needed to communicate with Genesys Cloud
func newIdpGsuiteProxy(clientConfig *platformclientv2.Configuration) *idpGsuiteProxy {
	api := platformclientv2.NewIdentityProviderApiWithConfig(clientConfig)
	return &idpGsuiteProxy{
		clientConfig:        clientConfig,
		identityProviderApi: api,
		getIdpGsuiteAttr:    getIdpGsuiteFn,
		updateIdpGsuiteAttr: updateIdpGsuiteFn,
		deleteIdpGsuiteAttr: deleteIdpGsuiteFn,
	}
}

// getIdpGsuiteProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getIdpGsuiteProxy(clientConfig *platformclientv2.Configuration) *idpGsuiteProxy {
	if internalProxy == nil {
		internalProxy = newIdpGsuiteProxy(clientConfig)
	}

	return internalProxy
}

// getIdpGsuite retrieves all Genesys Cloud idp gsuite
func (p *idpGsuiteProxy) getIdpGsuite(ctx context.Context) (*platformclientv2.Gsuite, *platformclientv2.APIResponse, error) {
	return p.getIdpGsuiteAttr(ctx, p)
}

// updateIdpGsuite updates a Genesys Cloud idp gsuite
func (p *idpGsuiteProxy) updateIdpGsuite(ctx context.Context, id string, idpGsuite *platformclientv2.Gsuite) (*platformclientv2.Identityprovider, *platformclientv2.APIResponse, error) {
	return p.updateIdpGsuiteAttr(ctx, p, id, idpGsuite)
}

// deleteIdpGsuite deletes a Genesys Cloud idp gsuite by Id
func (p *idpGsuiteProxy) deleteIdpGsuite(ctx context.Context, id string) (response *platformclientv2.APIResponse, err error) {
	return p.deleteIdpGsuiteAttr(ctx, p, id)
}

// getIdpGsuiteFn is the implementation for retrieving all idp gsuite in Genesys Cloud
func getIdpGsuiteFn(ctx context.Context, p *idpGsuiteProxy) (*platformclientv2.Gsuite, *platformclientv2.APIResponse, error) {
	return p.identityProviderApi.GetIdentityprovidersGsuite()
}

// updateIdpGsuiteFn is an implementation of the function to update a Genesys Cloud idp gsuite
func updateIdpGsuiteFn(ctx context.Context, p *idpGsuiteProxy, id string, idpGsuite *platformclientv2.Gsuite) (*platformclientv2.Identityprovider, *platformclientv2.APIResponse, error) {
	return p.identityProviderApi.PutIdentityprovidersGsuite(*idpGsuite)
}

// deleteIdpGsuiteFn is an implementation function for deleting a Genesys Cloud idp gsuite
func deleteIdpGsuiteFn(ctx context.Context, p *idpGsuiteProxy, id string) (response *platformclientv2.APIResponse, err error) {
	_, resp, err := p.identityProviderApi.DeleteIdentityprovidersGsuite()
	if err != nil {
		return resp, fmt.Errorf("Failed to delete idp gsuite: %s", err)
	}

	return resp, err
}
