package integration

import (
	"context"
	"fmt"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The genesyscloud_integration_proxy.go file contains the proxy structures and methods that interact
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
var internalProxy *integrationsProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getAllIntegrationsFunc func(ctx context.Context, p *integrationsProxy) (*[]platformclientv2.Integration, *platformclientv2.APIResponse, error)
type createIntegrationFunc func(ctx context.Context, p *integrationsProxy, integration *platformclientv2.Createintegrationrequest) (*platformclientv2.Integration, *platformclientv2.APIResponse, error)
type getIntegrationByIdFunc func(ctx context.Context, p *integrationsProxy, integrationId string) (integration *platformclientv2.Integration, response *platformclientv2.APIResponse, err error)
type getIntegrationByNameFunc func(ctx context.Context, p *integrationsProxy, integrationName string) (integration *platformclientv2.Integration, retryable bool, response *platformclientv2.APIResponse, err error)
type updateIntegrationFunc func(ctx context.Context, p *integrationsProxy, integrationId string, integration *platformclientv2.Integration) (*platformclientv2.Integration, *platformclientv2.APIResponse, error)
type deleteIntegrationFunc func(ctx context.Context, p *integrationsProxy, integrationId string) (response *platformclientv2.APIResponse, err error)
type getIntegrationConfigFunc func(ctx context.Context, p *integrationsProxy, integrationId string) (config *platformclientv2.Integrationconfiguration, response *platformclientv2.APIResponse, err error)
type updateIntegrationConfigFunc func(ctx context.Context, p *integrationsProxy, integrationId string, integrationConfig *platformclientv2.Integrationconfiguration) (integration *platformclientv2.Integrationconfiguration, response *platformclientv2.APIResponse, err error)

// integrationProxy contains all of the methods that call genesys cloud APIs.
type integrationsProxy struct {
	clientConfig                *platformclientv2.Configuration
	integrationsApi             *platformclientv2.IntegrationsApi
	getAllIntegrationsAttr      getAllIntegrationsFunc
	createIntegrationAttr       createIntegrationFunc
	getIntegrationByIdAttr      getIntegrationByIdFunc
	getIntegrationByNameAttr    getIntegrationByNameFunc
	updateIntegrationAttr       updateIntegrationFunc
	updateIntegrationConfigAttr updateIntegrationConfigFunc
	deleteIntegrationAttr       deleteIntegrationFunc
	getIntegrationConfigAttr    getIntegrationConfigFunc
}

// newIntegrationsProxy initializes the Integrations proxy with all of the data needed to communicate with Genesys Cloud
func newIntegrationsProxy(clientConfig *platformclientv2.Configuration) *integrationsProxy {
	api := platformclientv2.NewIntegrationsApiWithConfig(clientConfig)
	return &integrationsProxy{
		clientConfig:                clientConfig,
		integrationsApi:             api,
		getAllIntegrationsAttr:      getAllIntegrationsFn,
		createIntegrationAttr:       createIntegrationFn,
		getIntegrationByIdAttr:      getIntegrationByIdFn,
		getIntegrationByNameAttr:    getIntegrationByNameFn,
		updateIntegrationAttr:       updateIntegrationFn,
		updateIntegrationConfigAttr: updateIntegrationConfigFn,
		deleteIntegrationAttr:       deleteIntegrationFn,
		getIntegrationConfigAttr:    getIntegrationConfigFn,
	}
}

// getIntegrationsProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getIntegrationsProxy(clientConfig *platformclientv2.Configuration) *integrationsProxy {
	if internalProxy == nil {
		internalProxy = newIntegrationsProxy(clientConfig)
	}
	return internalProxy
}

// getAllIntegrations retrieves all Genesys Cloud Integrations
func (p *integrationsProxy) getAllIntegrations(ctx context.Context) (*[]platformclientv2.Integration, *platformclientv2.APIResponse, error) {
	return p.getAllIntegrationsAttr(ctx, p)
}

// createIntegration creates a Genesys Cloud Integration
func (p *integrationsProxy) createIntegration(ctx context.Context, integrationReq *platformclientv2.Createintegrationrequest) (*platformclientv2.Integration, *platformclientv2.APIResponse, error) {
	return p.createIntegrationAttr(ctx, p, integrationReq)
}

// getIntegrationById gets Genesys Cloud Integration by id
func (p *integrationsProxy) getIntegrationById(ctx context.Context, integrationId string) (*platformclientv2.Integration, *platformclientv2.APIResponse, error) {
	return p.getIntegrationByIdAttr(ctx, p, integrationId)
}

// getIntegrationByName gets a Genesys Cloud Integration by name
func (p *integrationsProxy) getIntegrationByName(ctx context.Context, integrationName string) (*platformclientv2.Integration, bool, *platformclientv2.APIResponse, error) {
	return p.getIntegrationByNameAttr(ctx, p, integrationName)
}

// updateIntegration updates a Genesys Cloud Integration
func (p *integrationsProxy) updateIntegration(ctx context.Context, integrationId string, integration *platformclientv2.Integration) (*platformclientv2.Integration, *platformclientv2.APIResponse, error) {
	return p.updateIntegrationAttr(ctx, p, integrationId, integration)
}

// deleteIntegration deletes a Genesys Cloud Integration
func (p *integrationsProxy) deleteIntegration(ctx context.Context, integrationId string) (response *platformclientv2.APIResponse, err error) {
	return p.deleteIntegrationAttr(ctx, p, integrationId)
}

// getIntegrationConfig get the current config of a Genesys Cloud Integration
func (p *integrationsProxy) getIntegrationConfig(ctx context.Context, integrationId string) (*platformclientv2.Integrationconfiguration, *platformclientv2.APIResponse, error) {
	return p.getIntegrationConfigAttr(ctx, p, integrationId)
}

// updateIntegrationConfig updates the config of a Genesys Cloud Integration
func (p *integrationsProxy) updateIntegrationConfig(ctx context.Context, integrationId string, integrationConfig *platformclientv2.Integrationconfiguration) (*platformclientv2.Integrationconfiguration, *platformclientv2.APIResponse, error) {
	return p.updateIntegrationConfigAttr(ctx, p, integrationId, integrationConfig)
}

// getAllIntegrationsFn is the implementation for retrieving all integrations in Genesys Cloud
func getAllIntegrationsFn(ctx context.Context, p *integrationsProxy) (*[]platformclientv2.Integration, *platformclientv2.APIResponse, error) {
	var allIntegrations []platformclientv2.Integration
	var resp *platformclientv2.APIResponse
	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		integrations, response, err := p.integrationsApi.GetIntegrations(pageSize, pageNum, "", nil, "", "")
		if err != nil {
			return nil, resp, err
		}
		resp = response
		if integrations.Entities == nil || len(*integrations.Entities) == 0 {
			break
		}
		allIntegrations = append(allIntegrations, *integrations.Entities...)
	}
	return &allIntegrations, resp, nil
}

// createIntegrationFn is the implementation for creating an integration in Genesys Cloud
func createIntegrationFn(ctx context.Context, p *integrationsProxy, integrationReq *platformclientv2.Createintegrationrequest) (*platformclientv2.Integration, *platformclientv2.APIResponse, error) {
	integration, resp, err := p.integrationsApi.PostIntegrations(*integrationReq)
	if err != nil {
		return nil, resp, err
	}
	return integration, resp, nil
}

// getIntegrationByIdFn is the implementation for getting a Genesys Cloud Integration by id
func getIntegrationByIdFn(ctx context.Context, p *integrationsProxy, integrationId string) (*platformclientv2.Integration, *platformclientv2.APIResponse, error) {
	const pageSize = 100
	const pageNum = 1
	integration, resp, err := p.integrationsApi.GetIntegration(integrationId, pageSize, pageNum, "", nil, "", "")
	if err != nil {
		return nil, resp, err
	}
	return integration, resp, nil
}

// getIntegrationByNameFn is the implementation for getting a Genesys Cloud Integration by name
func getIntegrationByNameFn(ctx context.Context, p *integrationsProxy, integrationName string) (*platformclientv2.Integration, bool, *platformclientv2.APIResponse, error) {
	var foundIntegration *platformclientv2.Integration
	var resp *platformclientv2.APIResponse
	const pageSize = 100
	for pageNum := 1; ; pageNum++ {
		integrations, response, err := p.integrationsApi.GetIntegrations(pageSize, pageNum, "", nil, "", "")
		if err != nil {
			return nil, false, resp, err
		}
		resp = response
		if integrations.Entities == nil || len(*integrations.Entities) == 0 {
			return nil, true, resp, fmt.Errorf("no integrations found with name: %s", integrationName)
		}

		for _, integration := range *integrations.Entities {
			if integration.Name != nil && *integration.Name == integrationName {
				foundIntegration = &integration
				break
			}
		}
		if foundIntegration != nil {
			break
		}
	}
	return foundIntegration, false, resp, nil
}

// updateIntegrationFn is the implementation for updating a Genesys Cloud Integration
func updateIntegrationFn(ctx context.Context, p *integrationsProxy, integrationId string, integration *platformclientv2.Integration) (*platformclientv2.Integration, *platformclientv2.APIResponse, error) {
	const pageSize = 25
	const pageNum = 1
	integration, resp, err := p.integrationsApi.PatchIntegration(integrationId, pageSize, pageNum, "", nil, "", "", *integration)
	if err != nil {
		return nil, resp, err
	}
	return integration, resp, nil
}

// deleteIntegrationFn is the implementation for deleting a Genesys Cloud Integration
func deleteIntegrationFn(ctx context.Context, p *integrationsProxy, integrationId string) (response *platformclientv2.APIResponse, err error) {
	_, resp, err := p.integrationsApi.DeleteIntegration(integrationId)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

// getIntegrationConfigFn is the implementation for getting the current config of a Genesys Cloud Integration
func getIntegrationConfigFn(ctx context.Context, p *integrationsProxy, integrationId string) (*platformclientv2.Integrationconfiguration, *platformclientv2.APIResponse, error) {
	config, resp, err := p.integrationsApi.GetIntegrationConfigCurrent(integrationId)
	if err != nil {
		return nil, resp, err
	}
	return config, resp, nil
}

// updateIntegrationConfigFn is the implementation for updating a Genesys Cloud Integration Config
func updateIntegrationConfigFn(ctx context.Context, p *integrationsProxy, integrationId string, integrationConfig *platformclientv2.Integrationconfiguration) (*platformclientv2.Integrationconfiguration, *platformclientv2.APIResponse, error) {
	config, resp, err := p.integrationsApi.PutIntegrationConfigCurrent(integrationId, *integrationConfig)
	if err != nil {
		return nil, resp, err
	}
	return config, resp, nil
}
