package integration_custom_auth_action

import (
	"context"
	"fmt"
	"strings"

	"github.com/mypurecloud/platform-client-sdk-go/v112/platformclientv2"
)

/*
The genesyscloud_integration_custom_auth_action_proxy.go file contains the proxy structures and methods that interact
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
var internalProxy *customAuthActionsProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getAllIntegrationCustomAuthActionsFunc func(ctx context.Context, p *customAuthActionsProxy) (*[]platformclientv2.Action, error)
type getCustomAuthActionByIdFunc func(ctx context.Context, p *customAuthActionsProxy, actionId string) (*platformclientv2.Action, *platformclientv2.APIResponse, error)
type updateCustomAuthActionFunc func(ctx context.Context, p *customAuthActionsProxy, actionId string, updateAction *platformclientv2.Updateactioninput) (*platformclientv2.Action, *platformclientv2.APIResponse, error)
type getIntegrationActionTemplateFunc func(ctx context.Context, p *customAuthActionsProxy, actionId string, fileName string) (*string, *platformclientv2.APIResponse, error)
type getIntegrationTypeFunc func(ctx context.Context, p *customAuthActionsProxy, integrationId string) (string, error)
type getIntegrationCredentialsTypeFunc func(ctx context.Context, p *customAuthActionsProxy, integrationId string) (string, error)
type getIntegrationByNameFunc func(ctx context.Context, p *customAuthActionsProxy, integrationName string) (integration *platformclientv2.Integration, retryable bool, err error)

// customAuthActionsProxy contains all of the methods that call genesys cloud APIs.
type customAuthActionsProxy struct {
	clientConfig                           *platformclientv2.Configuration
	integrationsApi                        *platformclientv2.IntegrationsApi
	getAllIntegrationCustomAuthActionsAttr getAllIntegrationCustomAuthActionsFunc
	getCustomAuthActionByIdAttr            getCustomAuthActionByIdFunc
	updateCustomAuthActionAttr             updateCustomAuthActionFunc
	getIntegrationActionTemplateAttr       getIntegrationActionTemplateFunc
	getIntegrationTypeAttr                 getIntegrationTypeFunc
	getIntegrationCredentialsTypeAttr      getIntegrationCredentialsTypeFunc
	getIntegrationByNameAttr               getIntegrationByNameFunc
}

// newCustomAuthActionsProxy initializes the customAuthActionsProxy with all of the data needed to communicate with Genesys Cloud
func newCustomAuthActionsProxy(clientConfig *platformclientv2.Configuration) *customAuthActionsProxy {
	api := platformclientv2.NewIntegrationsApiWithConfig(clientConfig)
	return &customAuthActionsProxy{
		clientConfig:                           clientConfig,
		integrationsApi:                        api,
		getAllIntegrationCustomAuthActionsAttr: getAllIntegrationCustomAuthActionsFn,
		getCustomAuthActionByIdAttr:            getCustomAuthActionByIdFn,
		updateCustomAuthActionAttr:             updateCustomAuthActionFn,
		getIntegrationActionTemplateAttr:       getIntegrationActionTemplateFn,
		getIntegrationTypeAttr:                 getIntegrationTypeFn,
		getIntegrationCredentialsTypeAttr:      getIntegrationCredentialsTypeFn,
		getIntegrationByNameAttr:               getIntegrationByNameFn,
	}
}

// getCustomAuthActionsProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getCustomAuthActionsProxy(clientConfig *platformclientv2.Configuration) *customAuthActionsProxy {
	if internalProxy == nil {
		internalProxy = newCustomAuthActionsProxy(clientConfig)
	}

	return internalProxy
}

// getAllIntegrationCustomAuthActions retrieves all Genesys Cloud Integration Custom Auth Actions
func (p *customAuthActionsProxy) getAllIntegrationCustomAuthActions(ctx context.Context) (*[]platformclientv2.Action, error) {
	return p.getAllIntegrationCustomAuthActionsAttr(ctx, p)
}

// getCustomAuthActionById retrieve a Genesys Cloud Integration Custom Auth Action by ID
func (p *customAuthActionsProxy) getCustomAuthActionById(ctx context.Context, actionId string) (*platformclientv2.Action, *platformclientv2.APIResponse, error) {
	return p.getCustomAuthActionByIdAttr(ctx, p, actionId)
}

// updateCustomAuthAction updates a Genesys Cloud Integration Custom Auth Action
func (p *customAuthActionsProxy) updateCustomAuthAction(ctx context.Context, actionId string, updateAction *platformclientv2.Updateactioninput) (*platformclientv2.Action, *platformclientv2.APIResponse, error) {
	return p.updateCustomAuthActionAttr(ctx, p, actionId, updateAction)
}

// getIntegrationActionTemplate retrieves a Genesys Cloud Integration Action Template
func (p *customAuthActionsProxy) getIntegrationActionTemplate(ctx context.Context, actionId string, fileName string) (*string, *platformclientv2.APIResponse, error) {
	return p.getIntegrationActionTemplateAttr(ctx, p, actionId, fileName)
}

// getIntegrationType retrieves the type of a Genesys Cloud Integration
func (p *customAuthActionsProxy) getIntegrationType(ctx context.Context, integrationId string) (string, error) {
	return p.getIntegrationTypeAttr(ctx, p, integrationId)
}

// getIntegrationCredentialsType retrieves the type of a Genesys Cloud Integration Credential
func (p *customAuthActionsProxy) getIntegrationCredentialsType(ctx context.Context, integrationId string) (string, error) {
	return p.getIntegrationCredentialsTypeAttr(ctx, p, integrationId)
}

// getIntegrationByName gets a Genesys Cloud Integration by name
func (p *customAuthActionsProxy) getIntegrationByName(ctx context.Context, integrationName string) (*platformclientv2.Integration, bool, error) {
	return p.getIntegrationByNameAttr(ctx, p, integrationName)
}

// getAllIntegrationCustomAuthActionsFn is the implementation for getting all integration custom auth actions in Genesys Cloud
func getAllIntegrationCustomAuthActionsFn(ctx context.Context, p *customAuthActionsProxy) (*[]platformclientv2.Action, error) {
	actions := []platformclientv2.Action{}

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		actionsList, _, err := p.integrationsApi.GetIntegrationsActions(pageSize, pageNum, "", "", "", "", "", "", "", "", "true")
		if err != nil {
			return nil, err
		}

		if actionsList.Entities == nil || len(*actionsList.Entities) == 0 {
			break
		}

		for _, action := range *actionsList.Entities {
			if !strings.HasPrefix(*action.Id, customAuthIdPrefix) {
				continue
			}
			actions = append(actions, action)
		}
	}

	return &actions, nil
}

// getCustomAuthActionByIdFn is the implementation for getting an integration custom auth actions by id in Genesys Cloud
func getCustomAuthActionByIdFn(ctx context.Context, p *customAuthActionsProxy, actionId string) (*platformclientv2.Action, *platformclientv2.APIResponse, error) {
	action, resp, err := p.integrationsApi.GetIntegrationsAction(actionId, "", true)
	if err != nil {
		return nil, resp, err
	}

	return action, resp, nil
}

// updateCustomAuthActionFn is the implementation for updating an integration custom auth action in Genesys Cloud
func updateCustomAuthActionFn(ctx context.Context, p *customAuthActionsProxy, actionId string, updateAction *platformclientv2.Updateactioninput) (*platformclientv2.Action, *platformclientv2.APIResponse, error) {
	action, resp, err := p.integrationsApi.PatchIntegrationsAction(actionId, *updateAction)
	if err != nil {
		return nil, resp, err
	}

	return action, resp, nil
}

// getIntegrationActionTemplateFn is the implementation for getting the integration action template in Genesys Cloud
func getIntegrationActionTemplateFn(ctx context.Context, p *customAuthActionsProxy, actionId string, fileName string) (*string, *platformclientv2.APIResponse, error) {
	template, resp, err := p.integrationsApi.GetIntegrationsActionTemplate(actionId, fileName)
	if err != nil {
		return nil, resp, err
	}
	return template, resp, nil
}

// getIntegrationTypeFn is the implementation for getting the type of an integration in Genesys Cloud
func getIntegrationTypeFn(ctx context.Context, p *customAuthActionsProxy, integrationId string) (string, error) {
	integration, _, err := p.integrationsApi.GetIntegration(integrationId, 1, 1, "", nil, "", "")
	if err != nil {
		return "", err
	}

	return *integration.IntegrationType.Id, nil
}

// getIntegrationCredentialsTypeFn is the implementation for getting the type of an integration credential in Genesys Cloud
func getIntegrationCredentialsTypeFn(ctx context.Context, p *customAuthActionsProxy, integrationId string) (string, error) {
	integrationConfig, _, err := p.integrationsApi.GetIntegrationConfigCurrent(integrationId)
	if err != nil {
		return "", err
	}
	if integrationConfig.Credentials == nil || len(*integrationConfig.Credentials) == 0 {
		return "", fmt.Errorf("no credentials set for integration %s", integrationId)
	}

	basicAuth, found := (*integrationConfig.Credentials)["basicAuth"]
	if !found {
		return "", fmt.Errorf("no 'basicAuth' credentials set for integration %s", integrationId)
	}

	credential, _, err := p.integrationsApi.GetIntegrationsCredential(*basicAuth.Id)
	if err != nil {
		return "", err
	}

	return *credential.VarType.Name, nil
}

// getIntegrationByNameFn is the implementation for getting a Genesys Cloud Integration by name
func getIntegrationByNameFn(ctx context.Context, p *customAuthActionsProxy, integrationName string) (*platformclientv2.Integration, bool, error) {
	var foundIntegration *platformclientv2.Integration

	const pageSize = 100
	for pageNum := 1; ; pageNum++ {
		integrations, _, err := p.integrationsApi.GetIntegrations(pageSize, pageNum, "", nil, "", "")
		if err != nil {
			return nil, false, err
		}

		if integrations.Entities == nil || len(*integrations.Entities) == 0 {
			return nil, true, fmt.Errorf("no integrations found with name: %s", integrationName)
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

	return foundIntegration, false, nil
}
