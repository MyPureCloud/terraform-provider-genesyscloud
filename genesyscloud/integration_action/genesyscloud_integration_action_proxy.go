package integration_action

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *integrationActionsProxy

type getAllIntegrationActionsFunc func(ctx context.Context, p *integrationActionsProxy) (*[]platformclientv2.Action, error)
type createIntegrationActionFunc func(ctx context.Context, p *integrationActionsProxy, action *IntegrationAction) (*IntegrationAction, *platformclientv2.APIResponse, error)
type getIntegrationActionByIdFunc func(ctx context.Context, p *integrationActionsProxy, actionId string) (*IntegrationAction, *platformclientv2.APIResponse, error)
type getIntegrationActionsByNameFunc func(ctx context.Context, p *integrationActionsProxy, actionName string) (actions *[]platformclientv2.Action, err error)
type updateIntegrationActionFunc func(ctx context.Context, p *integrationActionsProxy, actionId string, updateAction *platformclientv2.Updateactioninput) (*platformclientv2.Action, *platformclientv2.APIResponse, error)
type deleteIntegrationActionFunc func(ctx context.Context, p *integrationActionsProxy, actionId string) (*platformclientv2.APIResponse, error)
type getIntegrationActionTemplateFunc func(ctx context.Context, p *integrationActionsProxy, actionId string, fileName string) (*string, *platformclientv2.APIResponse, error)

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

// newIntegrationActionsProxy initializes the Integrations proxy with all of the data needed to communicate with Genesys Cloud
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

func (p *integrationActionsProxy) getAllIntegrationActions(ctx context.Context) (*[]platformclientv2.Action, error) {
	return p.getAllIntegrationActionsAttr(ctx, p)
}

func (p *integrationActionsProxy) createIntegrationAction(ctx context.Context, actionInput *IntegrationAction) (*IntegrationAction, *platformclientv2.APIResponse, error) {
	return p.createIntegrationActionAttr(ctx, p, actionInput)
}

func (p *integrationActionsProxy) getIntegrationActionById(ctx context.Context, actionId string) (action *IntegrationAction, response *platformclientv2.APIResponse, err error) {
	return p.getIntegrationActionByIdAttr(ctx, p, actionId)
}

func (p *integrationActionsProxy) getIntegrationActionsByName(ctx context.Context, actionName string) (actions *[]platformclientv2.Action, err error) {
	return p.getIntegrationActionsByNameAttr(ctx, p, actionName)
}

func (p *integrationActionsProxy) updateIntegrationAction(ctx context.Context, actionId string, updateAction *platformclientv2.Updateactioninput) (*platformclientv2.Action, *platformclientv2.APIResponse, error) {
	return p.updateIntegrationActionAttr(ctx, p, actionId, updateAction)
}

func (p *integrationActionsProxy) deleteIntegrationAction(ctx context.Context, actionId string) (*platformclientv2.APIResponse, error) {
	return p.deleteIntegrationActionAttr(ctx, p, actionId)
}

func (p *integrationActionsProxy) getIntegrationActionTemplate(ctx context.Context, actionId string, fileName string) (*string, *platformclientv2.APIResponse, error) {
	return p.getIntegrationActionTemplateAttr(ctx, p, actionId, fileName)
}

func getAllIntegrationActionsFn(ctx context.Context, p *integrationActionsProxy) (*[]platformclientv2.Action, error) {
	actions := []platformclientv2.Action{}

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		actionsList, _, err := p.integrationsApi.GetIntegrationsActions(pageSize, pageNum, "", "", "", "", "", "", "", "", "")
		if err != nil {
			return nil, err
		}

		if actionsList.Entities == nil || len(*actionsList.Entities) == 0 {
			break
		}

		actions = append(actions, *actionsList.Entities...)
	}

	return &actions, nil
}

func createIntegrationActionFn(ctx context.Context, p *integrationActionsProxy, actionInput *IntegrationAction) (*IntegrationAction, *platformclientv2.APIResponse, error) {
	action, resp, err := sdkPostIntegrationAction(actionInput, p.integrationsApi)
	if err != nil {
		return nil, resp, err
	}

	return action, resp, nil
}

func getIntegrationActionByIdFn(ctx context.Context, p *integrationActionsProxy, actionId string) (*IntegrationAction, *platformclientv2.APIResponse, error) {
	action, resp, err := sdkGetIntegrationAction(actionId, p.integrationsApi)
	if err != nil {
		return nil, resp, err
	}

	return action, resp, nil
}

// Get integration action by name.
func getIntegrationActionsByNameFn(ctx context.Context, p *integrationActionsProxy, actionName string) (*[]platformclientv2.Action, error) {
	var actions []platformclientv2.Action

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		integrationAction, _, err := p.integrationsApi.GetIntegrationsActions(pageSize, pageNum, "", "", "", "", "", actionName, "", "", "")
		if err != nil {
			return nil, err
		}

		if integrationAction.Entities == nil || len(*integrationAction.Entities) == 0 {
			break
		}

		for _, action := range *integrationAction.Entities {
			if action.Name != nil && *action.Name == actionName {
				actions = append(actions, action)
			}
		}
	}

	return &actions, nil
}

func updateIntegrationActionFn(ctx context.Context, p *integrationActionsProxy, actionId string, updateAction *platformclientv2.Updateactioninput) (*platformclientv2.Action, *platformclientv2.APIResponse, error) {
	action, resp, err := p.integrationsApi.PatchIntegrationsAction(actionId, *updateAction)
	if err != nil {
		return nil, resp, err
	}

	return action, resp, nil
}

func deleteIntegrationActionFn(ctx context.Context, p *integrationActionsProxy, actionId string) (*platformclientv2.APIResponse, error) {
	resp, err := p.integrationsApi.DeleteIntegrationsAction(actionId)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

func getIntegrationActionTemplateFn(ctx context.Context, p *integrationActionsProxy, actionId string, fileName string) (*string, *platformclientv2.APIResponse, error) {
	template, resp, err := sdkGetIntegrationActionTemplate(actionId, fileName, p.integrationsApi)
	if err != nil {
		return nil, resp, err
	}
	return template, resp, nil
}

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
