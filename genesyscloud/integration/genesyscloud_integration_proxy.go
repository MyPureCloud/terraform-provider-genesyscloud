package integration

import (
	"context"

	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

/*
The genesyscloud_externalcontacts_contact_proxy.go file contains the proxy structures and methods that interact
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
type getAllIntegrationsFunc func(ctx context.Context, p *integrationsProxy) (*[]platformclientv2.Integration, error)
type createIntegrationFunc func(ctx context.Context, p *integrationsProxy, integration *platformclientv2.Integration) (*platformclientv2.Integration, error)
type getIntegrationByIdFunc func(ctx context.Context, p *integrationsProxy, externalContactId string) (externalContact *platformclientv2.Externalcontact, responseCode int, err error)
type getIntegrationByNameFunc func(ctx context.Context, p *integrationsProxy, scriptName string) ([]platformclientv2.Script, error)
type updateIntegrationFunc func(ctx context.Context, p *integrationsProxy, externalContactId string, externalContact *platformclientv2.Externalcontact) (*platformclientv2.Externalcontact, error)
type deleteIntegrationFunc func(ctx context.Context, p *integrationsProxy, externalContactId string) (responseCode int, err error)

// integrationProxy contains all of the methods that call genesys cloud APIs.
type integrationsProxy struct {
	clientConfig             *platformclientv2.Configuration
	integrationsApi          *platformclientv2.IntegrationsApi
	getAllIntegrationsAttr   getAllIntegrationsFunc
	createIntegrationAttr    createIntegrationFunc
	getIntegrationByIdAttr   getIntegrationByIdFunc
	getIntegrationByNameAttr getIntegrationByNameFunc
	updateIntegrationAttr    updateIntegrationFunc
	deleteIntegrationAttr    deleteIntegrationFunc
}

// newIntegrationsProxy initializes the Integrations proxy with all of the data needed to communicate with Genesys Cloud
func newIntegrationsProxy(clientConfig *platformclientv2.Configuration) *integrationsProxy {
	api := platformclientv2.NewIntegrationsApiWithConfig(clientConfig)
	return &integrationsProxy{
		clientConfig:             clientConfig,
		integrationsApi:          api,
		getAllIntegrationsAttr:   getAllIntegrationsFn,
		createIntegrationAttr:    createIntegrationFn,
		getIntegrationByIdAttr:   getIntegrationByIdFn,
		getIntegrationByNameAttr: getIntegrationByNameFn,
		updateIntegrationAttr:    updateIntegrationFn,
		deleteIntegrationAttr:    deleteIntegrationFn,
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
func (p *integrationsProxy) getAllIntegrations(ctx context.Context) (*[]platformclientv2.Integration, error) {
	return p.getAllIntegrationsAttr(ctx, p)
}

// createIntegration creates a Genesys Cloud Integration
func (p *integrationsProxy) createIntegration(ctx context.Context, integration *platformclientv2.Integration) (*platformclientv2.Integration, error) {
	return p.createIntegrationAttr(ctx, p, integration)
}

// getAllIntegrationsFn is the implementation for retrieving all integrations in Genesys Cloud
func getAllIntegrationsFn(ctx context.Context, p *integrationsProxy) (*[]platformclientv2.Integration, error) {
	var allIntegrations []platformclientv2.Integration

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		integrations, _, err := p.integrationsApi.GetIntegrations(pageSize, pageNum, "", nil, "", "")
		if err != nil {
			return nil, err
		}

		if integrations.Entities == nil || len(*integrations.Entities) == 0 {
			break
		}

		for _, integration := range *integrations.Entities {
			allIntegrations = append(allIntegrations, integration)
		}
	}

	return &allIntegrations, nil
}

// createIntegrationFn is the implementation for creating an integration in Genesys Cloud
func createIntegrationFn(ctx context.Context, p *integrationsProxy, integration *platformclientv2.Integration) (*platformclientv2.Integration, error) {
	createIntegration := platformclientv2.Createintegrationrequest{
		IntegrationType: integration.IntegrationType,
	}

	integration, _, err := p.integrationsApi.PostIntegrations(createIntegration)
	if err != nil {
		return nil, err
	}
}
