package idp_adfs

import (
	"context"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The genesyscloud_idp_adfs_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *idpAdfsProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getAllIdpAdfsFunc func(ctx context.Context, p *idpAdfsProxy) (*platformclientv2.Adfs, *platformclientv2.APIResponse, error)
type updateIdpAdfsFunc func(ctx context.Context, p *idpAdfsProxy, id string, aDFS *platformclientv2.Adfs) (resp *platformclientv2.APIResponse, err error)
type deleteIdpAdfsFunc func(ctx context.Context, p *idpAdfsProxy, id string) (resp *platformclientv2.APIResponse, err error)

// idpAdfsProxy contains all of the methods that call genesys cloud APIs.
type idpAdfsProxy struct {
	clientConfig        *platformclientv2.Configuration
	identityProviderApi *platformclientv2.IdentityProviderApi
	getAllIdpAdfsAttr   getAllIdpAdfsFunc
	updateIdpAdfsAttr   updateIdpAdfsFunc
	deleteIdpAdfsAttr   deleteIdpAdfsFunc
}

// newIdpAdfsProxy initializes the idp adfs proxy with all of the data needed to communicate with Genesys Cloud
func newIdpAdfsProxy(clientConfig *platformclientv2.Configuration) *idpAdfsProxy {
	api := platformclientv2.NewIdentityProviderApiWithConfig(clientConfig)
	return &idpAdfsProxy{
		clientConfig:        clientConfig,
		identityProviderApi: api,
		getAllIdpAdfsAttr:   getAllIdpAdfsFn,
		updateIdpAdfsAttr:   updateIdpAdfsFn,
		deleteIdpAdfsAttr:   deleteIdpAdfsFn,
	}
}

// getIdpAdfsProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getIdpAdfsProxy(clientConfig *platformclientv2.Configuration) *idpAdfsProxy {
	if internalProxy == nil {
		internalProxy = newIdpAdfsProxy(clientConfig)
	}

	return internalProxy
}

// getIdpAdfs retrieves all Genesys Cloud idp adfs
func (p *idpAdfsProxy) getIdpAdfs(ctx context.Context) (*platformclientv2.Adfs, *platformclientv2.APIResponse, error) {
	return p.getAllIdpAdfsAttr(ctx, p)
}

// updateIdpAdfs updates a Genesys Cloud idp adfs
func (p *idpAdfsProxy) updateIdpAdfs(ctx context.Context, id string, idpAdfs *platformclientv2.Adfs) (resp *platformclientv2.APIResponse, err error) {
	return p.updateIdpAdfsAttr(ctx, p, id, idpAdfs)
}

// deleteIdpAdfs deletes a Genesys Cloud idp adfs by Id
func (p *idpAdfsProxy) deleteIdpAdfs(ctx context.Context, id string) (resp *platformclientv2.APIResponse, err error) {
	return p.deleteIdpAdfsAttr(ctx, p, id)
}

// getAllIdpAdfsFn is the implementation for retrieving all idp adfs in Genesys Cloud
func getAllIdpAdfsFn(ctx context.Context, p *idpAdfsProxy) (*platformclientv2.Adfs, *platformclientv2.APIResponse, error) {
	adfs, resp, err := p.identityProviderApi.GetIdentityprovidersAdfs()
	if err != nil {
		return nil, resp, err
	}

	return adfs, resp, nil
}

// updateIdpAdfsFn is an implementation of the function to update a Genesys Cloud idp adfs
func updateIdpAdfsFn(ctx context.Context, p *idpAdfsProxy, id string, idpAdfs *platformclientv2.Adfs) (statusCode *platformclientv2.APIResponse, err error) {
	_, resp, err := p.identityProviderApi.PutIdentityprovidersAdfs(*idpAdfs)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

// deleteIdpAdfsFn is an implementation function for deleting a Genesys Cloud idp adfs
func deleteIdpAdfsFn(ctx context.Context, p *idpAdfsProxy, id string) (response *platformclientv2.APIResponse, err error) {
	_, resp, err := p.identityProviderApi.DeleteIdentityprovidersAdfs()
	if err != nil {
		return resp, err
	}

	return resp, nil
}
