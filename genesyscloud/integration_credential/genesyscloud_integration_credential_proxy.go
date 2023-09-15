package integration_credential

import (
	"context"
	"fmt"

	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *integrationCredsProxy

type getAllIntegrationCredsFunc func(ctx context.Context, p *integrationCredsProxy) (*[]platformclientv2.Credentialinfo, error)
type createIntegrationCredFunc func(ctx context.Context, p *integrationCredsProxy, createCredential *platformclientv2.Credential) (*platformclientv2.Credentialinfo, error)
type getIntegrationCredByIdFunc func(ctx context.Context, p *integrationCredsProxy, credentialId string) (credential *platformclientv2.Credential, response *platformclientv2.APIResponse, err error)
type getIntegrationCredByNameFunc func(ctx context.Context, p *integrationCredsProxy, credentialName string) (*platformclientv2.Credentialinfo, error)
type updateIntegrationCredFunc func(ctx context.Context, p *integrationCredsProxy, credentialId string, credential *platformclientv2.Credential) (*platformclientv2.Credentialinfo, error)
type deleteIntegrationCredFunc func(ctx context.Context, p *integrationCredsProxy, credentialId string) (responseCode int, err error)

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
	}
}

// getIntegrationsProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getIntegrationCredsProxy(clientConfig *platformclientv2.Configuration) *integrationCredsProxy {
	if internalProxy == nil {
		internalProxy = newIntegrationCredsProxy(clientConfig)
	}

	return internalProxy
}

// getAllIntegrationCredentials retrieves all Genesys Cloud Integrations
func (p *integrationCredsProxy) getAllIntegrationCreds(ctx context.Context) (*[]platformclientv2.Credentialinfo, error) {
	return p.getAllIntegrationCredsAttr(ctx, p)
}

func (p *integrationCredsProxy) createIntegrationCred(ctx context.Context, createCredential *platformclientv2.Credential) (*platformclientv2.Credentialinfo, error) {
	return p.createIntegrationCredAttr(ctx, p, createCredential)
}

func (p *integrationCredsProxy) getIntegrationCredById(ctx context.Context, credentialId string) (credential *platformclientv2.Credential, response *platformclientv2.APIResponse, err error) {
	return p.getIntegrationCredByIdAttr(ctx, p, credentialId)
}

func (p *integrationCredsProxy) getIntegrationCredByName(ctx context.Context, credentialName string) (*platformclientv2.Credentialinfo, error) {
	return p.getIntegrationCredByNameAttr(ctx, p, credentialName)
}

func (p *integrationCredsProxy) updateIntegrationCred(ctx context.Context, credentialId string, credential *platformclientv2.Credential) (*platformclientv2.Credentialinfo, error) {
	return p.updateIntegrationCredAttr(ctx, p, credentialId, credential)
}

func (p *integrationCredsProxy) deleteIntegrationCred(ctx context.Context, credentialId string) (responseCode int, err error) {
	return p.deleteIntegrationCredAttr(ctx, p, credentialId)
}

func getAllIntegrationCredsFn(ctx context.Context, p *integrationCredsProxy) (*[]platformclientv2.Credentialinfo, error) {
	var allCreds []platformclientv2.Credentialinfo

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		credentials, _, err := p.integrationsApi.GetIntegrationsCredentials(pageNum, pageSize)
		if err != nil {
			return nil, err
		}

		if credentials.Entities == nil || len(*credentials.Entities) == 0 {
			break
		}

		allCreds = append(allCreds, *credentials.Entities...)
	}

	return &allCreds, nil
}

func createIntegrationCredFn(ctx context.Context, p *integrationCredsProxy, createCredential *platformclientv2.Credential) (*platformclientv2.Credentialinfo, error) {
	credential, _, err := p.integrationsApi.PostIntegrationsCredentials(*createCredential)
	if err != nil {
		return nil, err
	}

	return credential, nil
}

func getIntegrationCredByIdFn(ctx context.Context, p *integrationCredsProxy, credentialId string) (*platformclientv2.Credential, *platformclientv2.APIResponse, error) {
	credential, resp, err := p.integrationsApi.GetIntegrationsCredential(credentialId)
	if err != nil {
		return nil, resp, err
	}

	return credential, resp, nil
}

func getIntegrationCredByNameFn(ctx context.Context, p *integrationCredsProxy, credentialName string) (*platformclientv2.Credentialinfo, error) {
	var foundCred *platformclientv2.Credentialinfo

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		integrationCredentials, _, err := p.integrationsApi.GetIntegrationsCredentials(pageNum, pageSize)

		if err != nil {
			return nil, err
		}

		if integrationCredentials.Entities == nil || len(*integrationCredentials.Entities) == 0 {
			return nil, fmt.Errorf("no integration credentials found with name: %s", credentialName)
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

	return foundCred, nil
}

func updateIntegrationCredFn(ctx context.Context, p *integrationCredsProxy, credentialId string, credential *platformclientv2.Credential) (*platformclientv2.Credentialinfo, error) {
	credInfo, _, err := p.integrationsApi.PutIntegrationsCredential(credentialId, *credential)
	if err != nil {
		return nil, err
	}

	return credInfo, nil
}

func deleteIntegrationCredFn(ctx context.Context, p *integrationCredsProxy, credentialId string) (responseCode int, err error) {
	resp, err := p.integrationsApi.DeleteIntegrationsCredential(credentialId)
	if err != nil {
		return resp.StatusCode, err
	}

	return resp.StatusCode, nil
}
