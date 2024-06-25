package integration_credential

import (
	"context"
	"fmt"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The genesyscloud_integration_credential_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.

Each proxy implementation:

1.  Should provide a private package level variable that holds a instance of a proxy class.
2.  A New... constructor function  to initialize the proxy object.  This constructor should only be used within
    the proxy.
3.  A get private constructor function that the classes in the package can be used to to retrieve
    the proxy.  This proxy should check to see if the package level proxy instance is nil and
    should initialize it, otherwise it should return the instance
4.  Type definitions for each function that will be used in the proxy.  We use composition here
    so that we can easily provide mocks for testing.
5.  A struct for the proxy that holds an attribute for each function type.
6.  Wrapper methods on each of the elements on the struct.
7.  Function implementations for each function type definition.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *integrationCredsProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getAllIntegrationCredsFunc func(ctx context.Context, p *integrationCredsProxy) (*[]platformclientv2.Credentialinfo, *platformclientv2.APIResponse, error)
type createIntegrationCredFunc func(ctx context.Context, p *integrationCredsProxy, createCredential *platformclientv2.Credential) (*platformclientv2.Credentialinfo, *platformclientv2.APIResponse, error)
type getIntegrationCredByIdFunc func(ctx context.Context, p *integrationCredsProxy, credentialId string) (credential *platformclientv2.Credential, response *platformclientv2.APIResponse, err error)
type getIntegrationCredByNameFunc func(ctx context.Context, p *integrationCredsProxy, credentialName string) (credential *platformclientv2.Credentialinfo, retryable bool, response *platformclientv2.APIResponse, err error)
type updateIntegrationCredFunc func(ctx context.Context, p *integrationCredsProxy, credentialId string, credential *platformclientv2.Credential) (*platformclientv2.Credentialinfo, *platformclientv2.APIResponse, error)
type deleteIntegrationCredFunc func(ctx context.Context, p *integrationCredsProxy, credentialId string) (response *platformclientv2.APIResponse, err error)
type getIntegrationByIdFunc func(ctx context.Context, p *integrationCredsProxy, integrationId string) (integration *platformclientv2.Integration, response *platformclientv2.APIResponse, err error)

// integrationCredsProxy contains all of the methods that call genesys cloud APIs.
type integrationCredsProxy struct {
	clientConfig                 *platformclientv2.Configuration
	integrationsApi              *platformclientv2.IntegrationsApi
	getAllIntegrationCredsAttr   getAllIntegrationCredsFunc
	createIntegrationCredAttr    createIntegrationCredFunc
	getIntegrationCredByIdAttr   getIntegrationCredByIdFunc
	getIntegrationCredByNameAttr getIntegrationCredByNameFunc
	updateIntegrationCredAttr    updateIntegrationCredFunc
	deleteIntegrationCredAttr    deleteIntegrationCredFunc
	getIntegrationByIdAttr       getIntegrationByIdFunc
}

// newIntegrationCredsProxy initializes the Integration Credentials proxy with all of the data needed to communicate with Genesys Cloud
func newIntegrationCredsProxy(clientConfig *platformclientv2.Configuration) *integrationCredsProxy {
	api := platformclientv2.NewIntegrationsApiWithConfig(clientConfig)
	return &integrationCredsProxy{
		clientConfig:                 clientConfig,
		integrationsApi:              api,
		getAllIntegrationCredsAttr:   getAllIntegrationCredsFn,
		createIntegrationCredAttr:    createIntegrationCredFn,
		getIntegrationCredByIdAttr:   getIntegrationCredByIdFn,
		getIntegrationCredByNameAttr: getIntegrationCredByNameFn,
		updateIntegrationCredAttr:    updateIntegrationCredFn,
		deleteIntegrationCredAttr:    deleteIntegrationCredFn,
		getIntegrationByIdAttr:       getIntegrationByIdFn,
	}
}

// getIntegrationCredsProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getIntegrationCredsProxy(clientConfig *platformclientv2.Configuration) *integrationCredsProxy {
	if internalProxy == nil {
		internalProxy = newIntegrationCredsProxy(clientConfig)
	}
	return internalProxy
}

// getAllIntegrationCredentials retrieves all Genesys Cloud Integrations
func (p *integrationCredsProxy) getAllIntegrationCreds(ctx context.Context) (*[]platformclientv2.Credentialinfo, *platformclientv2.APIResponse, error) {
	return p.getAllIntegrationCredsAttr(ctx, p)
}

// createIntegrationCred creates a Genesys Cloud Crdential
func (p *integrationCredsProxy) createIntegrationCred(ctx context.Context, createCredential *platformclientv2.Credential) (*platformclientv2.Credentialinfo, *platformclientv2.APIResponse, error) {
	return p.createIntegrationCredAttr(ctx, p, createCredential)
}

// getIntegrationCredById gets a Genesys Cloud Integration Credential by id
func (p *integrationCredsProxy) getIntegrationCredById(ctx context.Context, credentialId string) (credential *platformclientv2.Credential, response *platformclientv2.APIResponse, err error) {
	return p.getIntegrationCredByIdAttr(ctx, p, credentialId)
}

// getIntegrationCredByName gets a Genesys Cloud Integration Credential by name
func (p *integrationCredsProxy) getIntegrationCredByName(ctx context.Context, credentialName string) (*platformclientv2.Credentialinfo, bool, *platformclientv2.APIResponse, error) {
	return p.getIntegrationCredByNameAttr(ctx, p, credentialName)
}

// updateIntegrationCred udpates a Genesys Cloud Integration Credential
func (p *integrationCredsProxy) updateIntegrationCred(ctx context.Context, credentialId string, credential *platformclientv2.Credential) (*platformclientv2.Credentialinfo, *platformclientv2.APIResponse, error) {
	return p.updateIntegrationCredAttr(ctx, p, credentialId, credential)
}

// deleteIntegrationCred deletes a Genesys Cloud Integration Credential
func (p *integrationCredsProxy) deleteIntegrationCred(ctx context.Context, credentialId string) (response *platformclientv2.APIResponse, err error) {
	return p.deleteIntegrationCredAttr(ctx, p, credentialId)
}

// getIntegrationById gets Genesys Cloud Integration by id
func (p *integrationCredsProxy) getIntegrationById(ctx context.Context, integrationId string) (*platformclientv2.Integration, *platformclientv2.APIResponse, error) {
	return p.getIntegrationByIdAttr(ctx, p, integrationId)
}

// getAllIntegrationCredsFn is the implementation for getting all integration credentials in Genesys Cloud
func getAllIntegrationCredsFn(ctx context.Context, p *integrationCredsProxy) (*[]platformclientv2.Credentialinfo, *platformclientv2.APIResponse, error) {
	var allCreds []platformclientv2.Credentialinfo
	var resp *platformclientv2.APIResponse
	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		credentials, response, err := p.integrationsApi.GetIntegrationsCredentials(pageNum, pageSize)
		if err != nil {
			return nil, response, err
		}
		resp = response
		if credentials.Entities == nil || len(*credentials.Entities) == 0 {
			break
		}

		allCreds = append(allCreds, *credentials.Entities...)
	}
	return &allCreds, resp, nil
}

// createIntegrationCredFn is the implementation for creating an integration credential in Genesys Cloud
func createIntegrationCredFn(ctx context.Context, p *integrationCredsProxy, createCredential *platformclientv2.Credential) (*platformclientv2.Credentialinfo, *platformclientv2.APIResponse, error) {
	credential, resp, err := p.integrationsApi.PostIntegrationsCredentials(*createCredential)
	if err != nil {
		return nil, resp, err
	}
	return credential, resp, nil
}

// getIntegrationCredByIdFn is the implementation for getting an integration credential by id in Genesys Cloud
func getIntegrationCredByIdFn(ctx context.Context, p *integrationCredsProxy, credentialId string) (*platformclientv2.Credential, *platformclientv2.APIResponse, error) {
	credential, resp, err := p.integrationsApi.GetIntegrationsCredential(credentialId)
	if err != nil {
		return nil, resp, err
	}
	return credential, resp, nil
}

// getIntegrationCredByNameFn is the implementation for getting an integration credential by name in Genesys Cloud
func getIntegrationCredByNameFn(ctx context.Context, p *integrationCredsProxy, credentialName string) (*platformclientv2.Credentialinfo, bool, *platformclientv2.APIResponse, error) {
	var foundCred *platformclientv2.Credentialinfo
	var resp *platformclientv2.APIResponse
	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		integrationCredentials, response, err := p.integrationsApi.GetIntegrationsCredentials(pageNum, pageSize)
		resp = response
		if err != nil {
			return nil, false, response, err
		}

		if integrationCredentials.Entities == nil || len(*integrationCredentials.Entities) == 0 {
			return nil, true, response, fmt.Errorf("no integration credentials found with name: %s", credentialName)
		}

		for _, credential := range *integrationCredentials.Entities {
			if credential.Name != nil && *credential.Name == credentialName {
				foundCred = &credential
				break
			}
		}
		if foundCred != nil {
			break
		}
	}
	return foundCred, false, resp, nil
}

// updateIntegrationCredFn is the implementation for updating an integration credential in Genesys Cloud
func updateIntegrationCredFn(ctx context.Context, p *integrationCredsProxy, credentialId string, credential *platformclientv2.Credential) (*platformclientv2.Credentialinfo, *platformclientv2.APIResponse, error) {
	credInfo, resp, err := p.integrationsApi.PutIntegrationsCredential(credentialId, *credential)
	if err != nil {
		return nil, resp, err
	}
	return credInfo, resp, nil
}

// deleteIntegrationCredFn is the implementation for deleting an integration credential in Genesys Cloud
func deleteIntegrationCredFn(ctx context.Context, p *integrationCredsProxy, credentialId string) (response *platformclientv2.APIResponse, err error) {
	resp, err := p.integrationsApi.DeleteIntegrationsCredential(credentialId)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

// getIntegrationByIdFn is the implementation for getting a Genesys Cloud Integration by id
func getIntegrationByIdFn(ctx context.Context, p *integrationCredsProxy, integrationId string) (*platformclientv2.Integration, *platformclientv2.APIResponse, error) {
	const pageSize = 100
	const pageNum = 1
	integration, resp, err := p.integrationsApi.GetIntegration(integrationId, pageSize, pageNum, "", nil, "", "")
	if err != nil {
		return nil, resp, err
	}
	return integration, resp, nil
}
