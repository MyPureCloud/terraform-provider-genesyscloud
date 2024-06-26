package idp_ping

import (
	"context"
	"fmt"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The genesyscloud_idp_ping_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *idpPingProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getIdpPingFunc func(ctx context.Context, p *idpPingProxy) (*platformclientv2.Pingidentity, *platformclientv2.APIResponse, error)
type updateIdpPingFunc func(ctx context.Context, p *idpPingProxy, id string, pingIdentity *platformclientv2.Pingidentity) (*platformclientv2.Identityprovider, *platformclientv2.APIResponse, error)
type deleteIdpPingFunc func(ctx context.Context, p *idpPingProxy, id string) (response *platformclientv2.APIResponse, err error)

// idpPingProxy contains all of the methods that call genesys cloud APIs.
type idpPingProxy struct {
	clientConfig        *platformclientv2.Configuration
	identityProviderApi *platformclientv2.IdentityProviderApi
	getIdpPingAttr      getIdpPingFunc
	updateIdpPingAttr   updateIdpPingFunc
	deleteIdpPingAttr   deleteIdpPingFunc
}

// newIdpPingProxy initializes the idp ping proxy with all of the data needed to communicate with Genesys Cloud
func newIdpPingProxy(clientConfig *platformclientv2.Configuration) *idpPingProxy {
	api := platformclientv2.NewIdentityProviderApiWithConfig(clientConfig)
	return &idpPingProxy{
		clientConfig:        clientConfig,
		identityProviderApi: api,
		getIdpPingAttr:      getIdpPingFn,
		updateIdpPingAttr:   updateIdpPingFn,
		deleteIdpPingAttr:   deleteIdpPingFn,
	}
}

// getIdpPingProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getIdpPingProxy(clientConfig *platformclientv2.Configuration) *idpPingProxy {
	if internalProxy == nil {
		internalProxy = newIdpPingProxy(clientConfig)
	}

	return internalProxy
}

// getIdpPing retrieves all Genesys Cloud idp ping
func (p *idpPingProxy) getIdpPing(ctx context.Context) (*platformclientv2.Pingidentity, *platformclientv2.APIResponse, error) {
	return p.getIdpPingAttr(ctx, p)
}

// updateIdpPing updates a Genesys Cloud idp ping
func (p *idpPingProxy) updateIdpPing(ctx context.Context, id string, idpPing *platformclientv2.Pingidentity) (*platformclientv2.Identityprovider, *platformclientv2.APIResponse, error) {
	return p.updateIdpPingAttr(ctx, p, id, idpPing)
}

// deleteIdpPing deletes a Genesys Cloud idp ping by Id
func (p *idpPingProxy) deleteIdpPing(ctx context.Context, id string) (response *platformclientv2.APIResponse, err error) {
	return p.deleteIdpPingAttr(ctx, p, id)
}

// getAllIdpPingFn is the implementation for retrieving all idp ping in Genesys Cloud
func getIdpPingFn(ctx context.Context, p *idpPingProxy) (*platformclientv2.Pingidentity, *platformclientv2.APIResponse, error) {
	return p.identityProviderApi.GetIdentityprovidersPing()
}

// updateIdpPingFn is an implementation of the function to update a Genesys Cloud idp ping
func updateIdpPingFn(ctx context.Context, p *idpPingProxy, id string, idpPing *platformclientv2.Pingidentity) (*platformclientv2.Identityprovider, *platformclientv2.APIResponse, error) {
	return p.identityProviderApi.PutIdentityprovidersPing(*idpPing)
}

// deleteIdpPingFn is an implementation function for deleting a Genesys Cloud idp ping
func deleteIdpPingFn(ctx context.Context, p *idpPingProxy, id string) (response *platformclientv2.APIResponse, err error) {
	_, resp, err := p.identityProviderApi.DeleteIdentityprovidersPing()
	if err != nil {
		return resp, fmt.Errorf("Failed to delete idp ping: %s", err)
	}

	return resp, err
}
