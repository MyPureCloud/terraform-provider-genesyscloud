package integration_action

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The genesyscloud_integration_action_proxy.go file contains the proxy structures and methods that interact
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

/*
NOTE: Most of the integration action methods invokes the API manually instead of using the Genesys Cloud Go SDK types
and API methods. This is due to the limitation of the output contract.
In the SDK the input and output contracts are of the Jsonschemadocument type. This defines a JSON schema
for the contract. The type has the usual properties like 'Name' and 'Properties' however it is missing the 'Items'
property which is needed to define the item type of an array.
In the API, the output contract allows the root to be of 'array' type instead of 'object'. If that is the case it requires
the 'Items' property to define the 'object' schema it allows. Since it's impossible to do with the SDK,
helper methods and types are created to invoke the APIs with Genesys Cloud.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *integrationActionsProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getAllIntegrationActionsFunc func(ctx context.Context, p *integrationActionsProxy) (*[]platformclientv2.Action, *platformclientv2.APIResponse, error)
type createIntegrationActionFunc func(ctx context.Context, p *integrationActionsProxy, action *IntegrationAction) (*IntegrationAction, *platformclientv2.APIResponse, error)
type getIntegrationActionByIdFunc func(ctx context.Context, p *integrationActionsProxy, actionId string) (*IntegrationAction, *platformclientv2.APIResponse, error)
type getIntegrationActionsByNameFunc func(ctx context.Context, p *integrationActionsProxy, actionName string) (actions *[]platformclientv2.Action, response *platformclientv2.APIResponse, err error)
type updateIntegrationActionFunc func(ctx context.Context, p *integrationActionsProxy, actionId string, updateAction *platformclientv2.Updateactioninput) (*platformclientv2.Action, *platformclientv2.APIResponse, error)
type deleteIntegrationActionFunc func(ctx context.Context, p *integrationActionsProxy, actionId string) (*platformclientv2.APIResponse, error)
type getIntegrationActionTemplateFunc func(ctx context.Context, p *integrationActionsProxy, actionId string, fileName string) (*string, *platformclientv2.APIResponse, error)

// integrationActionsProxy contains all of the methods that call genesys cloud APIs.
type integrationActionsProxy struct {
	clientConfig                     *platformclientv2.Configuration
	integrationsApi                  *platformclientv2.IntegrationsApi
	getAllIntegrationActionsAttr     getAllIntegrationActionsFunc
	createIntegrationActionAttr      createIntegrationActionFunc
	getIntegrationActionByIdAttr     getIntegrationActionByIdFunc
	getIntegrationActionsByNameAttr  getIntegrationActionsByNameFunc
	updateIntegrationActionAttr      updateIntegrationActionFunc
	deleteIntegrationActionAttr      deleteIntegrationActionFunc
	getIntegrationActionTemplateAttr getIntegrationActionTemplateFunc
}

// newIntegrationActionsProxy initializes the integrationActionsProxy with all of the data needed to communicate with Genesys Cloud
func newIntegrationActionsProxy(clientConfig *platformclientv2.Configuration) *integrationActionsProxy {
	api := platformclientv2.NewIntegrationsApiWithConfig(clientConfig)
	return &integrationActionsProxy{
		clientConfig:                     clientConfig,
		integrationsApi:                  api,
		getAllIntegrationActionsAttr:     getAllIntegrationActionsFn,
		createIntegrationActionAttr:      createIntegrationActionFn,
		getIntegrationActionByIdAttr:     getIntegrationActionByIdFn,
		getIntegrationActionsByNameAttr:  getIntegrationActionsByNameFn,
		updateIntegrationActionAttr:      updateIntegrationActionFn,
		deleteIntegrationActionAttr:      deleteIntegrationActionFn,
		getIntegrationActionTemplateAttr: getIntegrationActionTemplateFn,
	}
}

// getIntegrationActionsProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getIntegrationActionsProxy(clientConfig *platformclientv2.Configuration) *integrationActionsProxy {
	if internalProxy == nil {
		internalProxy = newIntegrationActionsProxy(clientConfig)
	}
	return internalProxy
}

// getAllIntegrationActions retrieves all Genesys Cloud Integration Actions
func (p *integrationActionsProxy) getAllIntegrationActions(ctx context.Context) (*[]platformclientv2.Action, *platformclientv2.APIResponse, error) {
	return p.getAllIntegrationActionsAttr(ctx, p)
}

// createIntegrationAction creates a Genesys Cloud Integration Action
func (p *integrationActionsProxy) createIntegrationAction(ctx context.Context, actionInput *IntegrationAction) (*IntegrationAction, *platformclientv2.APIResponse, error) {
	return p.createIntegrationActionAttr(ctx, p, actionInput)
}

// getIntegrationActionById gets a Genesys Cloud Integration Action by id
func (p *integrationActionsProxy) getIntegrationActionById(ctx context.Context, actionId string) (action *IntegrationAction, response *platformclientv2.APIResponse, err error) {
	return p.getIntegrationActionByIdAttr(ctx, p, actionId)
}

// getIntegrationActionsByName gets a Genesys Cloud Integration Action by name
func (p *integrationActionsProxy) getIntegrationActionsByName(ctx context.Context, actionName string) (actions *[]platformclientv2.Action, response *platformclientv2.APIResponse, err error) {
	return p.getIntegrationActionsByNameAttr(ctx, p, actionName)
}

// updateIntegrationAction updates a Genesys Cloud Integration Action
func (p *integrationActionsProxy) updateIntegrationAction(ctx context.Context, actionId string, updateAction *platformclientv2.Updateactioninput) (*platformclientv2.Action, *platformclientv2.APIResponse, error) {
	return p.updateIntegrationActionAttr(ctx, p, actionId, updateAction)
}

// deleteIntegrationAction deletes a Genesys Cloud Integration Action
func (p *integrationActionsProxy) deleteIntegrationAction(ctx context.Context, actionId string) (*platformclientv2.APIResponse, error) {
	return p.deleteIntegrationActionAttr(ctx, p, actionId)
}

// getIntegrationActionTemplate gets a Genesys Cloud Integration Action Contract Template by its filename
func (p *integrationActionsProxy) getIntegrationActionTemplate(ctx context.Context, actionId string, fileName string) (*string, *platformclientv2.APIResponse, error) {
	return p.getIntegrationActionTemplateAttr(ctx, p, actionId, fileName)
}

// getAllIntegrationActionsFn is the implementation for retrieving all integration actions in Genesys Cloud
func getAllIntegrationActionsFn(ctx context.Context, p *integrationActionsProxy) (*[]platformclientv2.Action, *platformclientv2.APIResponse, error) {
	actions := []platformclientv2.Action{}
	var resp *platformclientv2.APIResponse
	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		actionsList, response, err := p.integrationsApi.GetIntegrationsActions(pageSize, pageNum, "", "", "", "", "", "", "", "", "")
		if err != nil {
			return nil, resp, err
		}
		resp = response
		if actionsList.Entities == nil || len(*actionsList.Entities) == 0 {
			break
		}
		actions = append(actions, *actionsList.Entities...)
	}
	return &actions, resp, nil
}

// createIntegrationActionFn is the implementation for creating an integration action in Genesys Cloud
func createIntegrationActionFn(ctx context.Context, p *integrationActionsProxy, actionInput *IntegrationAction) (*IntegrationAction, *platformclientv2.APIResponse, error) {
	action, resp, err := sdkPostIntegrationAction(actionInput, p.integrationsApi)
	if err != nil {
		return nil, resp, err
	}
	return action, resp, nil
}

// getIntegrationActionByIdFn is the implementation for getting an integration action by id in Genesys Cloud
func getIntegrationActionByIdFn(ctx context.Context, p *integrationActionsProxy, actionId string) (*IntegrationAction, *platformclientv2.APIResponse, error) {
	action, resp, err := sdkGetIntegrationAction(actionId, p.integrationsApi)
	if err != nil {
		return nil, resp, err
	}
	return action, resp, nil
}

// getIntegrationActionsByNameFn is the implementation for getting an integration action by name in Genesys Cloud
func getIntegrationActionsByNameFn(ctx context.Context, p *integrationActionsProxy, actionName string) (*[]platformclientv2.Action, *platformclientv2.APIResponse, error) {
	var actions []platformclientv2.Action
	var resp *platformclientv2.APIResponse
	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		integrationAction, response, err := p.integrationsApi.GetIntegrationsActions(pageSize, pageNum, "", "", "", "", "", actionName, "", "", "")
		if err != nil {
			return nil, response, err
		}
		resp = response
		if integrationAction.Entities == nil || len(*integrationAction.Entities) == 0 {
			break
		}

		for _, action := range *integrationAction.Entities {
			if action.Name != nil && *action.Name == actionName {
				actions = append(actions, action)
			}
		}
	}
	return &actions, resp, nil
}

// updateIntegrationActionFn is the implementation for updating an integration action in Genesys Cloud
func updateIntegrationActionFn(ctx context.Context, p *integrationActionsProxy, actionId string, updateAction *platformclientv2.Updateactioninput) (*platformclientv2.Action, *platformclientv2.APIResponse, error) {
	action, resp, err := p.integrationsApi.PatchIntegrationsAction(actionId, *updateAction)
	if err != nil {
		return nil, resp, err
	}
	return action, resp, nil
}

// deleteIntegrationActionFn is the implementation for deleting an integration action in Genesys Cloud
func deleteIntegrationActionFn(ctx context.Context, p *integrationActionsProxy, actionId string) (*platformclientv2.APIResponse, error) {
	resp, err := p.integrationsApi.DeleteIntegrationsAction(actionId)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

// getIntegrationActionTemplateFn is the implementation for getting the integration action template in Genesys Cloud
func getIntegrationActionTemplateFn(ctx context.Context, p *integrationActionsProxy, actionId string, fileName string) (*string, *platformclientv2.APIResponse, error) {
	template, resp, err := sdkGetIntegrationActionTemplate(actionId, fileName, p.integrationsApi)
	if err != nil {
		return nil, resp, err
	}
	return template, resp, nil
}

// sdkPostIntegrationAction is the non-sdk helper method for creating an Integration Action
func sdkPostIntegrationAction(body *IntegrationAction, api *platformclientv2.IntegrationsApi) (*IntegrationAction, *platformclientv2.APIResponse, error) {
	apiClient := &api.Configuration.APIClient

	// create path and map variables
	path := api.Configuration.BasePath + "/api/v2/integrations/actions"

	headerParams := make(map[string]string)

	// add default headers if any
	for key := range api.Configuration.DefaultHeader {
		headerParams[key] = api.Configuration.DefaultHeader[key]
	}

	headerParams["Authorization"] = "Bearer " + api.Configuration.AccessToken
	headerParams["Content-Type"] = "application/json"
	headerParams["Accept"] = "application/json"

	var successPayload *IntegrationAction
	response, err := apiClient.CallAPI(path, http.MethodPost, body, headerParams, nil, nil, "", nil)
	if err != nil {
		// Nothing special to do here, but do avoid processing the response
	} else if response.Error != nil {
		err = errors.New(response.ErrorMessage)
	} else {
		err = json.Unmarshal([]byte(response.RawBody), &successPayload)
	}
	return successPayload, response, err
}

// sdkGetIntegrationAction is the non-sdk helper method for getting an Integration Action
func sdkGetIntegrationAction(actionId string, api *platformclientv2.IntegrationsApi) (*IntegrationAction, *platformclientv2.APIResponse, error) {
	apiClient := &api.Configuration.APIClient

	// create path and map variables
	path := api.Configuration.BasePath + "/api/v2/integrations/actions/" + actionId

	headerParams := make(map[string]string)
	queryParams := make(map[string]string)

	// oauth required
	if api.Configuration.AccessToken != "" {
		headerParams["Authorization"] = "Bearer " + api.Configuration.AccessToken
	}
	// add default headers if any
	for key := range api.Configuration.DefaultHeader {
		headerParams[key] = api.Configuration.DefaultHeader[key]
	}

	queryParams["expand"] = "contract"
	queryParams["includeConfig"] = "true"

	headerParams["Content-Type"] = "application/json"
	headerParams["Accept"] = "application/json"

	var successPayload *IntegrationAction
	response, err := apiClient.CallAPI(path, http.MethodGet, nil, headerParams, queryParams, nil, "", nil)
	if err != nil {
		// Nothing special to do here, but do avoid processing the response
	} else if response.Error != nil {
		err = errors.New(response.ErrorMessage)
	} else {
		err = json.Unmarshal([]byte(response.RawBody), &successPayload)
	}
	return successPayload, response, err
}

// sdkGetIntegrationActionTemplate is the non-sdk helper method for getting an Integration Action Template
func sdkGetIntegrationActionTemplate(actionId, templateName string, api *platformclientv2.IntegrationsApi) (*string, *platformclientv2.APIResponse, error) {
	apiClient := &api.Configuration.APIClient

	// create path and map variables
	path := api.Configuration.BasePath + "/api/v2/integrations/actions/" + actionId + "/templates/" + templateName

	headerParams := make(map[string]string)
	queryParams := make(map[string]string)

	// oauth required
	if api.Configuration.AccessToken != "" {
		headerParams["Authorization"] = "Bearer " + api.Configuration.AccessToken
	}
	// add default headers if any
	for key := range api.Configuration.DefaultHeader {
		headerParams[key] = api.Configuration.DefaultHeader[key]
	}

	headerParams["Content-Type"] = "application/json"
	headerParams["Accept"] = "*/*"

	var successPayload *string
	response, err := apiClient.CallAPI(path, http.MethodGet, nil, headerParams, queryParams, nil, "", nil)
	if err != nil {
		// Nothing special to do here, but do avoid processing the response
	} else if response.Error != nil {
		err = errors.New(response.ErrorMessage)
	} else {
		templateStr := string(response.RawBody)
		successPayload = &templateStr
	}
	return successPayload, response, err
}
