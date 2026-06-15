package integration_credential

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	rc "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v191/platformclientv2"
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
var internalProxy *IntegrationCredsProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getAllIntegrationCredsFunc func(ctx context.Context, p *IntegrationCredsProxy) (*[]platformclientv2.Credentialinfo, *platformclientv2.APIResponse, error)
type createIntegrationCredFunc func(ctx context.Context, p *IntegrationCredsProxy, createCredential *platformclientv2.Credential) (*platformclientv2.Credentialinfo, *platformclientv2.APIResponse, error)
type getIntegrationCredByIdFunc func(ctx context.Context, p *IntegrationCredsProxy, credentialId string) (credential *platformclientv2.Credential, response *platformclientv2.APIResponse, err error)
type getIntegrationCredByNameFunc func(ctx context.Context, p *IntegrationCredsProxy, credentialName string) (credential *platformclientv2.Credentialinfo, retryable bool, response *platformclientv2.APIResponse, err error)
type updateIntegrationCredFunc func(ctx context.Context, p *IntegrationCredsProxy, credentialId string, credential *platformclientv2.Credential) (*platformclientv2.Credentialinfo, *platformclientv2.APIResponse, error)
type deleteIntegrationCredFunc func(ctx context.Context, p *IntegrationCredsProxy, credentialId string) (response *platformclientv2.APIResponse, err error)
type getIntegrationByCredentialIdFunc func(ctx context.Context, p *IntegrationCredsProxy, integrationId string) (integration *platformclientv2.Integration, response *platformclientv2.APIResponse, err error)

// IntegrationCredsProxy contains all of the methods that call genesys cloud APIs.
type IntegrationCredsProxy struct {
	clientConfig                     *platformclientv2.Configuration
	integrationsApi                  *platformclientv2.IntegrationsApi
	getAllIntegrationCredsAttr       getAllIntegrationCredsFunc
	createIntegrationCredAttr        createIntegrationCredFunc
	getIntegrationCredByIdAttr       getIntegrationCredByIdFunc
	getIntegrationCredByNameAttr     getIntegrationCredByNameFunc
	updateIntegrationCredAttr        updateIntegrationCredFunc
	deleteIntegrationCredAttr        deleteIntegrationCredFunc
	getIntegrationByCredentialIdAttr getIntegrationByCredentialIdFunc
	integrationCache                 rc.CacheInterface[platformclientv2.Integration]
}

// newIntegrationCredsProxy initializes the Integration Credentials proxy with all of the data needed to communicate with Genesys Cloud
func newIntegrationCredsProxy(clientConfig *platformclientv2.Configuration) *IntegrationCredsProxy {
	api := platformclientv2.NewIntegrationsApiWithConfig(clientConfig)
	return &IntegrationCredsProxy{
		clientConfig:                     clientConfig,
		integrationsApi:                  api,
		getAllIntegrationCredsAttr:       getAllIntegrationCredsFn,
		createIntegrationCredAttr:        createIntegrationCredFn,
		getIntegrationCredByIdAttr:       getIntegrationCredByIdFn,
		getIntegrationCredByNameAttr:     getIntegrationCredByNameFn,
		updateIntegrationCredAttr:        updateIntegrationCredFn,
		deleteIntegrationCredAttr:        deleteIntegrationCredFn,
		getIntegrationByCredentialIdAttr: getIntegrationByCredentialIdFn,
		integrationCache:                 rc.NewResourceCache[platformclientv2.Integration](),
	}
}

// getIntegrationCredsProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getIntegrationCredsProxy(clientConfig *platformclientv2.Configuration) *IntegrationCredsProxy {
	if internalProxy == nil {
		internalProxy = newIntegrationCredsProxy(clientConfig)
	}
	return internalProxy
}

// GetIntegrationCredsProxy returns the proxy instance for use by other packages (e.g. tests)
func GetIntegrationCredsProxy(clientConfig *platformclientv2.Configuration) *IntegrationCredsProxy {
	return getIntegrationCredsProxy(clientConfig)
}

// getAllIntegrationCreds retrieves all Genesys Cloud Integration Credentials using cursor-based paging
func (p *IntegrationCredsProxy) getAllIntegrationCreds(ctx context.Context) (*[]platformclientv2.Credentialinfo, *platformclientv2.APIResponse, error) {
	return p.getAllIntegrationCredsAttr(ctx, p)
}

// createIntegrationCred creates a Genesys Cloud Crdential
func (p *IntegrationCredsProxy) createIntegrationCred(ctx context.Context, createCredential *platformclientv2.Credential) (*platformclientv2.Credentialinfo, *platformclientv2.APIResponse, error) {
	return p.createIntegrationCredAttr(ctx, p, createCredential)
}

// getIntegrationCredById gets a Genesys Cloud Integration Credential by id
func (p *IntegrationCredsProxy) getIntegrationCredById(ctx context.Context, credentialId string) (credential *platformclientv2.Credential, response *platformclientv2.APIResponse, err error) {
	return p.getIntegrationCredByIdAttr(ctx, p, credentialId)
}

// getIntegrationCredByName gets a Genesys Cloud Integration Credential by name
func (p *IntegrationCredsProxy) getIntegrationCredByName(ctx context.Context, credentialName string) (*platformclientv2.Credentialinfo, bool, *platformclientv2.APIResponse, error) {
	return p.getIntegrationCredByNameAttr(ctx, p, credentialName)
}

// updateIntegrationCred updates a Genesys Cloud Integration Credential
func (p *IntegrationCredsProxy) updateIntegrationCred(ctx context.Context, credentialId string, credential *platformclientv2.Credential) (*platformclientv2.Credentialinfo, *platformclientv2.APIResponse, error) {
	return p.updateIntegrationCredAttr(ctx, p, credentialId, credential)
}

// deleteIntegrationCred deletes a Genesys Cloud Integration Credential
func (p *IntegrationCredsProxy) deleteIntegrationCred(ctx context.Context, credentialId string) (response *platformclientv2.APIResponse, err error) {
	return p.deleteIntegrationCredAttr(ctx, p, credentialId)
}

// GetIntegrationByCredentialId is the public wrapper for getting the Genesys Cloud Integration by id
func (p *IntegrationCredsProxy) GetIntegrationByCredentialId(ctx context.Context, credentialId string) (*platformclientv2.Integration, *platformclientv2.APIResponse, error) {
	return p.getIntegrationByCredentialIdAttr(ctx, p, credentialId)
}

// getAllIntegrationCredsFn is the implementation for getting all integration credentials in Genesys Cloud using cursor-based paging
func getAllIntegrationCredsFn(ctx context.Context, p *IntegrationCredsProxy) (*[]platformclientv2.Credentialinfo, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	var allCreds []platformclientv2.Credentialinfo
	var after string
	pageSize := "200"
	var lastResp *platformclientv2.APIResponse

	for {
		credentials, resp, err := p.integrationsApi.GetIntegrationsCredentialsListing("", after, pageSize)
		if err != nil {
			return nil, resp, err
		}
		lastResp = resp

		if credentials.Entities != nil {
			allCreds = append(allCreds, *credentials.Entities...)
		}

		// Check if there are more pages
		if credentials.Entities == nil || len(*credentials.Entities) == 0 {
			break
		}

		// Use the last item's ID as the cursor for the next page
		lastItem := (*credentials.Entities)[len(*credentials.Entities)-1]
		if lastItem.Id == nil {
			break
		}
		after = *lastItem.Id
	}

	return &allCreds, lastResp, nil
}

// createIntegrationCredFn is the implementation for creating an integration credential in Genesys Cloud
func createIntegrationCredFn(ctx context.Context, p *IntegrationCredsProxy, createCredential *platformclientv2.Credential) (*platformclientv2.Credentialinfo, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	credential, resp, err := p.integrationsApi.PostIntegrationsCredentials(*createCredential)
	if err != nil {
		return nil, resp, err
	}
	return credential, resp, nil
}

// getIntegrationCredByIdFn is the implementation for getting an integration credential by id in Genesys Cloud
func getIntegrationCredByIdFn(ctx context.Context, p *IntegrationCredsProxy, credentialId string) (*platformclientv2.Credential, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	credential, resp, err := p.integrationsApi.GetIntegrationsCredential(credentialId)
	if err != nil {
		return nil, resp, err
	}
	return credential, resp, nil
}

// getIntegrationCredByNameFn is the implementation for getting an integration credential by name in Genesys Cloud
func getIntegrationCredByNameFn(ctx context.Context, p *IntegrationCredsProxy, credentialName string) (*platformclientv2.Credentialinfo, bool, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	var foundCred *platformclientv2.Credentialinfo
	var resp *platformclientv2.APIResponse
	var after string
	pageSize := "200"

	for {
		integrationCredentials, response, err := p.integrationsApi.GetIntegrationsCredentialsListing("", after, pageSize)
		resp = response
		if err != nil {
			return nil, false, response, err
		}

		if integrationCredentials.Entities == nil || len(*integrationCredentials.Entities) == 0 {
			break
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

		// Use the last item's ID as the cursor for the next page
		lastItem := (*integrationCredentials.Entities)[len(*integrationCredentials.Entities)-1]
		if lastItem.Id == nil {
			break
		}
		after = *lastItem.Id
	}

	if foundCred == nil {
		return nil, true, resp, fmt.Errorf("no integration credentials found with name: %s", credentialName)
	}

	return foundCred, false, resp, nil
}

// updateIntegrationCredFn is the implementation for updating an integration credential in Genesys Cloud
func updateIntegrationCredFn(ctx context.Context, p *IntegrationCredsProxy, credentialId string, credential *platformclientv2.Credential) (*platformclientv2.Credentialinfo, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	credInfo, resp, err := p.integrationsApi.PutIntegrationsCredential(credentialId, *credential)
	if err != nil {
		return nil, resp, err
	}
	return credInfo, resp, nil
}

// deleteIntegrationCredFn is the implementation for deleting an integration credential in Genesys Cloud
func deleteIntegrationCredFn(ctx context.Context, p *IntegrationCredsProxy, credentialId string) (response *platformclientv2.APIResponse, err error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	resp, err := p.integrationsApi.DeleteIntegrationsCredential(credentialId)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

// getIntegrationByCredentialIdFn is the implementation for getting a Genesys Cloud Integration by credential id.
// It builds a credential-to-integration cache on first call to avoid repeated API calls during export.
func getIntegrationByCredentialIdFn(ctx context.Context, p *IntegrationCredsProxy, credentialId string) (*platformclientv2.Integration, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	// Check cache first
	if cached := rc.GetCacheItem(p.integrationCache, credentialId); cached != nil {
		return cached, nil, nil
	}

	// Cache miss — fetch all integrations and their configs, then populate cache
	var allIntegrations []platformclientv2.Integration
	const pageSize = 100
	for pageNum := 1; ; pageNum++ {
		integrations, response, err := p.integrationsApi.GetIntegrations(pageSize, pageNum, "", nil, "", "", nil, "", "")
		if err != nil {
			return nil, response, err
		}
		if integrations.Entities == nil || len(*integrations.Entities) == 0 {
			break
		}
		allIntegrations = append(allIntegrations, *integrations.Entities...)
	}

	var foundIntegration platformclientv2.Integration
	var foundResp *platformclientv2.APIResponse
	for _, integ := range allIntegrations {
		if integ.Id == nil || *integ.Id == "" {
			continue
		}
		integrationConfig, response, err := p.integrationsApi.GetIntegrationConfigCurrent(*integ.Id)
		if err != nil {
			tflog.Warn(ctx, "Failed to retrieve integration config for integration "+*integ.Id, map[string]interface{}{"error": err, "response": response})
			continue
		}
		if integrationConfig.Credentials == nil {
			continue
		}
		for _, cred := range *integrationConfig.Credentials {
			if cred.Id != nil {
				rc.SetCache(p.integrationCache, *cred.Id, integ)
				if *cred.Id == credentialId {
					foundResp = response
					foundIntegration = integ
				}
			}
		}
	}
	if foundIntegration.Id != nil {
		return &foundIntegration, foundResp, nil
	}

	return nil, nil, fmt.Errorf("failed to find integration using credential id: %s", credentialId)
}
