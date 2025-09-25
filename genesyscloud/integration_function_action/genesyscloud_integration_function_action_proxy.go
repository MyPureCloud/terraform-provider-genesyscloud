package integration_function_action

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The genesyscloud_integration_function_action_proxy.go file contains the proxy structures and methods that interact
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
var internalProxy *integrationFunctionActionsProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getAllIntegrationFunctionActionsFunc func(ctx context.Context, p *integrationFunctionActionsProxy) (*[]platformclientv2.Action, *platformclientv2.APIResponse, error)
type createIntegrationFunctionActionFunc func(ctx context.Context, p *integrationFunctionActionsProxy, action *IntegrationFunctionAction) (*IntegrationFunctionAction, *platformclientv2.APIResponse, error)
type getIntegrationFunctionActionByIdFunc func(ctx context.Context, p *integrationFunctionActionsProxy, actionId string) (*IntegrationFunctionAction, *platformclientv2.APIResponse, error)
type getIntegrationFunctionActionsByNameFunc func(ctx context.Context, p *integrationFunctionActionsProxy, actionName string) (actions *[]platformclientv2.Action, response *platformclientv2.APIResponse, err error)
type updateIntegrationFunctionActionFunc func(ctx context.Context, p *integrationFunctionActionsProxy, actionId string, updateAction *platformclientv2.Updateactioninput) (*platformclientv2.Action, *platformclientv2.APIResponse, error)
type deleteIntegrationFunctionActionFunc func(ctx context.Context, p *integrationFunctionActionsProxy, actionId string) (*platformclientv2.APIResponse, error)
type getIntegrationFunctionActionTemplateFunc func(ctx context.Context, p *integrationFunctionActionsProxy, actionId string, fileName string) (*string, *platformclientv2.APIResponse, error)

// integrationFunctionActionsProxy contains all of the methods that call genesys cloud APIs.
type integrationFunctionActionsProxy struct {
	clientConfig                             *platformclientv2.Configuration
	integrationsApi                          *platformclientv2.IntegrationsApi
	getAllIntegrationFunctionActionsAttr     getAllIntegrationFunctionActionsFunc
	createIntegrationFunctionActionAttr      createIntegrationFunctionActionFunc
	getIntegrationFunctionActionByIdAttr     getIntegrationFunctionActionByIdFunc
	getIntegrationFunctionActionsByNameAttr  getIntegrationFunctionActionsByNameFunc
	updateIntegrationFunctionActionAttr      updateIntegrationFunctionActionFunc
	deleteIntegrationFunctionActionAttr      deleteIntegrationFunctionActionFunc
	getIntegrationFunctionActionTemplateAttr getIntegrationFunctionActionTemplateFunc
}

// newIntegrationFunctionActionsProxy initializes the IntegrationFunctionActionsProxy with all of the data needed to communicate with Genesys Cloud
func newIntegrationFunctionActionsProxy(clientConfig *platformclientv2.Configuration) *integrationFunctionActionsProxy {
	api := platformclientv2.NewIntegrationsApiWithConfig(clientConfig)
	return &integrationFunctionActionsProxy{
		clientConfig:                             clientConfig,
		integrationsApi:                          api,
		getAllIntegrationFunctionActionsAttr:     getAllIntegrationFunctionActionsFn,
		createIntegrationFunctionActionAttr:      createIntegrationFunctionActionFn,
		getIntegrationFunctionActionByIdAttr:     getIntegrationFunctionActionByIdFn,
		getIntegrationFunctionActionsByNameAttr:  getIntegrationFunctionActionsByNameFn,
		updateIntegrationFunctionActionAttr:      updateIntegrationFunctionActionFn,
		deleteIntegrationFunctionActionAttr:      deleteIntegrationFunctionActionFn,
		getIntegrationFunctionActionTemplateAttr: getIntegrationFunctionActionTemplateFn,
	}
}

// getIntegrationFunctionActionsProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getIntegrationFunctionActionsProxy(clientConfig *platformclientv2.Configuration) *integrationFunctionActionsProxy {
	if internalProxy == nil {
		internalProxy = newIntegrationFunctionActionsProxy(clientConfig)
	}
	return internalProxy
}

// getAllIntegrationFunctionActions retrieves all Genesys Cloud Integration Actions
func (p *integrationFunctionActionsProxy) getAllIntegrationFunctionActions(ctx context.Context) (*[]platformclientv2.Action, *platformclientv2.APIResponse, error) {
	return p.getAllIntegrationFunctionActionsAttr(ctx, p)
}

// createIntegrationFunctionAction creates a Genesys Cloud Integration Action
func (p *integrationFunctionActionsProxy) createIntegrationFunctionAction(ctx context.Context, actionInput *IntegrationFunctionAction) (*IntegrationFunctionAction, *platformclientv2.APIResponse, error) {
	return p.createIntegrationFunctionActionAttr(ctx, p, actionInput)
}

// getIntegrationFunctionActionById gets a Genesys Cloud Integration Action by id
func (p *integrationFunctionActionsProxy) getIntegrationFunctionActionById(ctx context.Context, actionId string) (action *IntegrationFunctionAction, response *platformclientv2.APIResponse, err error) {
	return p.getIntegrationFunctionActionByIdAttr(ctx, p, actionId)
}

// getIntegrationFunctionActionsByName gets a Genesys Cloud Integration Action by name
func (p *integrationFunctionActionsProxy) getIntegrationFunctionActionsByName(ctx context.Context, actionName string) (actions *[]platformclientv2.Action, response *platformclientv2.APIResponse, err error) {
	return p.getIntegrationFunctionActionsByNameAttr(ctx, p, actionName)
}

// updateIntegrationFunctionAction updates a Genesys Cloud Integration Action
func (p *integrationFunctionActionsProxy) updateIntegrationFunctionAction(ctx context.Context, actionId string, updateAction *platformclientv2.Updateactioninput) (*platformclientv2.Action, *platformclientv2.APIResponse, error) {
	return p.updateIntegrationFunctionActionAttr(ctx, p, actionId, updateAction)
}

// deleteIntegrationFunctionAction deletes a Genesys Cloud Integration Action
func (p *integrationFunctionActionsProxy) deleteIntegrationFunctionAction(ctx context.Context, actionId string) (*platformclientv2.APIResponse, error) {
	return p.deleteIntegrationFunctionActionAttr(ctx, p, actionId)
}

// getIntegrationFunctionActionTemplate gets a Genesys Cloud Integration Action Contract Template by its filename
func (p *integrationFunctionActionsProxy) getIntegrationFunctionActionTemplate(ctx context.Context, actionId string, fileName string) (*string, *platformclientv2.APIResponse, error) {
	return p.getIntegrationFunctionActionTemplateAttr(ctx, p, actionId, fileName)
}

// getAllIntegrationFunctionActionsFn is the implementation for retrieving all integration actions in Genesys Cloud
func getAllIntegrationFunctionActionsFn(ctx context.Context, p *integrationFunctionActionsProxy) (*[]platformclientv2.Action, *platformclientv2.APIResponse, error) {
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

// createIntegrationFunctionActionFn is the implementation for creating an integration action in Genesys Cloud
func createIntegrationFunctionActionFn(ctx context.Context, p *integrationFunctionActionsProxy, actionInput *IntegrationFunctionAction) (*IntegrationFunctionAction, *platformclientv2.APIResponse, error) {
	action, resp, err := sdkPostIntegrationFunctionAction(actionInput, p.integrationsApi)
	if err != nil {
		return nil, resp, err
	}
	return action, resp, nil
}

// getIntegrationFunctionActionByIdFn is the implementation for getting an integration action by id in Genesys Cloud
func getIntegrationFunctionActionByIdFn(ctx context.Context, p *integrationFunctionActionsProxy, actionId string) (*IntegrationFunctionAction, *platformclientv2.APIResponse, error) {
	action, resp, err := sdkGetIntegrationFunctionAction(actionId, p.integrationsApi)
	if err != nil {
		return nil, resp, err
	}
	return action, resp, nil
}

// getIntegrationFunctionActionsByNameFn is the implementation for getting an integration action by name in Genesys Cloud
func getIntegrationFunctionActionsByNameFn(ctx context.Context, p *integrationFunctionActionsProxy, actionName string) (*[]platformclientv2.Action, *platformclientv2.APIResponse, error) {
	var actions []platformclientv2.Action
	var resp *platformclientv2.APIResponse
	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		integrationFunctionAction, response, err := p.integrationsApi.GetIntegrationsActions(pageSize, pageNum, "", "", "", "", "", actionName, "", "", "")
		if err != nil {
			return nil, response, err
		}
		resp = response
		if integrationFunctionAction.Entities == nil || len(*integrationFunctionAction.Entities) == 0 {
			break
		}

		for _, action := range *integrationFunctionAction.Entities {
			if action.Name != nil && *action.Name == actionName {
				actions = append(actions, action)
			}
		}
	}
	return &actions, resp, nil
}

// updateIntegrationFunctionActionFn is the implementation for updating an integration action in Genesys Cloud
func updateIntegrationFunctionActionFn(ctx context.Context, p *integrationFunctionActionsProxy, actionId string, updateAction *platformclientv2.Updateactioninput) (*platformclientv2.Action, *platformclientv2.APIResponse, error) {
	action, resp, err := p.integrationsApi.PatchIntegrationsAction(actionId, *updateAction)
	if err != nil {
		return nil, resp, err
	}
	return action, resp, nil
}

// deleteIntegrationActionFn is the implementation for deleting an integration action in Genesys Cloud
func deleteIntegrationFunctionActionFn(ctx context.Context, p *integrationFunctionActionsProxy, actionId string) (*platformclientv2.APIResponse, error) {
	resp, err := p.integrationsApi.DeleteIntegrationsAction(actionId)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

// getIntegrationActionTemplateFn is the implementation for getting the integration action template in Genesys Cloud
func getIntegrationFunctionActionTemplateFn(ctx context.Context, p *integrationFunctionActionsProxy, actionId string, fileName string) (*string, *platformclientv2.APIResponse, error) {
	template, resp, err := sdkGetIntegrationFunctionActionTemplate(actionId, fileName, p.integrationsApi)
	if err != nil {
		return nil, resp, err
	}
	return template, resp, nil
}

// sdkPostIntegrationFunctionAction is the non-sdk helper method for creating an Integration Action
func sdkPostIntegrationFunctionAction(body *IntegrationFunctionAction, api *platformclientv2.IntegrationsApi) (*IntegrationFunctionAction, *platformclientv2.APIResponse, error) {
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

	var successPayload *IntegrationFunctionAction
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

// sdkGetIntegrationFunctionAction is the non-sdk helper method for getting an Integration Action
func sdkGetIntegrationFunctionAction(actionId string, api *platformclientv2.IntegrationsApi) (*IntegrationFunctionAction, *platformclientv2.APIResponse, error) {
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

	var successPayload *IntegrationFunctionAction
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

// sdkGetIntegrationFunctionActionTemplate is the non-sdk helper method for getting an Integration Action Template
func sdkGetIntegrationFunctionActionTemplate(actionId, templateName string, api *platformclientv2.IntegrationsApi) (*string, *platformclientv2.APIResponse, error) {
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
