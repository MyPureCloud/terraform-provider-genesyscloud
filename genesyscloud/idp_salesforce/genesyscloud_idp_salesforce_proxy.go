package idp_salesforce

import (
	"context"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The genesyscloud_idp_salesforce_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *idpSalesforceProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getIdpSalesforceFunc func(ctx context.Context, p *idpSalesforceProxy) (salesforce *platformclientv2.Salesforce, resp *platformclientv2.APIResponse, err error)
type updateIdpSalesforceFunc func(ctx context.Context, p *idpSalesforceProxy, salesforce *platformclientv2.Salesforce) (*platformclientv2.Identityprovider, *platformclientv2.APIResponse, error)
type deleteIdpSalesforceFunc func(ctx context.Context, p *idpSalesforceProxy) (resp *platformclientv2.APIResponse, err error)

// idpSalesforceProxy contains all of the methods that call genesys cloud APIs.
type idpSalesforceProxy struct {
	clientConfig            *platformclientv2.Configuration
	identityProviderApi     *platformclientv2.IdentityProviderApi
	getIdpSalesforceAttr    getIdpSalesforceFunc
	updateIdpSalesforceAttr updateIdpSalesforceFunc
	deleteIdpSalesforceAttr deleteIdpSalesforceFunc
}

// newIdpSalesforceProxy initializes the idp salesforce proxy with all of the data needed to communicate with Genesys Cloud
func newIdpSalesforceProxy(clientConfig *platformclientv2.Configuration) *idpSalesforceProxy {
	api := platformclientv2.NewIdentityProviderApiWithConfig(clientConfig)
	return &idpSalesforceProxy{
		clientConfig:            clientConfig,
		identityProviderApi:     api,
		getIdpSalesforceAttr:    getIdpSalesforceFn,
		updateIdpSalesforceAttr: updateIdpSalesforceFn,
		deleteIdpSalesforceAttr: deleteIdpSalesforceFn,
	}
}

// getIdpSalesforceProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getIdpSalesforceProxy(clientConfig *platformclientv2.Configuration) *idpSalesforceProxy {
	if internalProxy == nil {
		internalProxy = newIdpSalesforceProxy(clientConfig)
	}

	return internalProxy
}

// getIdpSalesforce returns a single Genesys Cloud idp salesforce
func (p *idpSalesforceProxy) getIdpSalesforce(ctx context.Context) (idpSalesforce *platformclientv2.Salesforce, resp *platformclientv2.APIResponse, err error) {
	return p.getIdpSalesforceAttr(ctx, p)
}

// updateIdpSalesforce updates a Genesys Cloud idp salesforce
func (p *idpSalesforceProxy) updateIdpSalesforce(ctx context.Context, idpSalesforce *platformclientv2.Salesforce) (*platformclientv2.Identityprovider, *platformclientv2.APIResponse, error) {
	return p.updateIdpSalesforceAttr(ctx, p, idpSalesforce)
}

// deleteIdpSalesforce deletes a Genesys Cloud idp salesforce by Id
func (p *idpSalesforceProxy) deleteIdpSalesforce(ctx context.Context) (resp *platformclientv2.APIResponse, err error) {
	return p.deleteIdpSalesforceAttr(ctx, p)
}

// getIdpSalesforceFn is an implementation of the function to get a Genesys Cloud idp salesforce
func getIdpSalesforceFn(ctx context.Context, p *idpSalesforceProxy) (idpSalesforce *platformclientv2.Salesforce, resp *platformclientv2.APIResponse, err error) {
	salesforce, resp, err := p.identityProviderApi.GetIdentityprovidersSalesforce()
	if err != nil {
		return nil, resp, err
	}

	return salesforce, resp, err
}

// updateIdpSalesforceFn is an implementation of the function to update a Genesys Cloud idp salesforce
func updateIdpSalesforceFn(ctx context.Context, p *idpSalesforceProxy, idpSalesforce *platformclientv2.Salesforce) (*platformclientv2.Identityprovider, *platformclientv2.APIResponse, error) {
	salesForce, resp, err := p.identityProviderApi.PutIdentityprovidersSalesforce(*idpSalesforce)
	if err != nil {
		return nil, resp, err
	}

	return salesForce, resp, nil
}

// deleteIdpSalesforceFn is an implementation function for deleting a Genesys Cloud idp salesforce
func deleteIdpSalesforceFn(ctx context.Context, p *idpSalesforceProxy) (resp *platformclientv2.APIResponse, err error) {
	_, resp, err = p.identityProviderApi.DeleteIdentityprovidersSalesforce()
	return resp, err
}
