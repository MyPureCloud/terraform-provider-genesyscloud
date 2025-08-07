package integration_action_draft

import (
	"context"
	"fmt"

	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

var internalProxy *integrationActionsProxy

type getAllIntegrationActionDraftsFunc func(ctx context.Context, p *integrationActionsProxy, name string) (*[]platformclientv2.Action, *platformclientv2.APIResponse, error)
type createIntegrationActionDraftFunc func(ctx context.Context, p *integrationActionsProxy, body platformclientv2.Postactioninput) (*platformclientv2.Action, *platformclientv2.APIResponse, error)
type getIntegrationActionDraftByIdFunc func(ctx context.Context, p *integrationActionsProxy, actionId string) (*platformclientv2.Action, *platformclientv2.APIResponse, error)
type getIntegrationActionDraftByNameFunc func(ctx context.Context, p *integrationActionsProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error)
type updateIntegrationActionDraftFunc func(ctx context.Context, p *integrationActionsProxy, actionId string, body platformclientv2.Updatedraftinput) (*platformclientv2.Action, *platformclientv2.APIResponse, error)
type deleteIntegrationActionDraftFunc func(ctx context.Context, p *integrationActionsProxy, actionId string) (*platformclientv2.APIResponse, error)
type getIntegrationActionDraftTemplateFunc func(ctx context.Context, p *integrationActionsProxy, actionId string, fileName string) (*string, *platformclientv2.APIResponse, error)

type integrationActionsProxy struct {
	clientConfig    *platformclientv2.Configuration
	integrationsApi *platformclientv2.IntegrationsApi

	getAllIntegrationActionDraftsAttr     getAllIntegrationActionDraftsFunc
	createIntegrationActionDraftAttr      createIntegrationActionDraftFunc
	getIntegrationActionDraftByIdAttr     getIntegrationActionDraftByIdFunc
	getIntegrationActionDraftByNameAttr   getIntegrationActionDraftByNameFunc
	updateIntegrationActionDraftAttr      updateIntegrationActionDraftFunc
	deleteIntegrationActionDraftAttr      deleteIntegrationActionDraftFunc
	getIntegrationActionDraftTemplateAttr getIntegrationActionDraftTemplateFunc
}

// newIntegrationActionsProxy initializes the integrationActionsProxy with all of the data needed to communicate with Genesys Cloud
func newIntegrationActionsProxy(clientConfig *platformclientv2.Configuration) *integrationActionsProxy {
	api := platformclientv2.NewIntegrationsApiWithConfig(clientConfig)
	return &integrationActionsProxy{
		clientConfig:                          clientConfig,
		integrationsApi:                       api,
		getAllIntegrationActionDraftsAttr:     getAllIntegrationActionDraftsFn,
		createIntegrationActionDraftAttr:      createIntegrationActionDraftFn,
		getIntegrationActionDraftByIdAttr:     getIntegrationActionDraftByIdFn,
		getIntegrationActionDraftByNameAttr:   getIntegrationActionDraftByNameFn,
		updateIntegrationActionDraftAttr:      updateIntegrationActionDraftFn,
		deleteIntegrationActionDraftAttr:      deleteIntegrationActionDraftFn,
		getIntegrationActionDraftTemplateAttr: getIntegrationActionDraftTemplateFn,
	}
}

func getIntegrationActionsProxy(clientConfig *platformclientv2.Configuration) *integrationActionsProxy {
	if internalProxy == nil {
		internalProxy = newIntegrationActionsProxy(clientConfig)
	}
	return internalProxy
}

func (p *integrationActionsProxy) getAllIntegrationActionDrafts(ctx context.Context, name string) (*[]platformclientv2.Action, *platformclientv2.APIResponse, error) {
	return p.getAllIntegrationActionDraftsAttr(ctx, p, name)
}

func (p *integrationActionsProxy) createIntegrationActionDraft(ctx context.Context, body platformclientv2.Postactioninput) (*platformclientv2.Action, *platformclientv2.APIResponse, error) {
	return p.createIntegrationActionDraftAttr(ctx, p, body)
}

func (p *integrationActionsProxy) getIntegrationActionDraftById(ctx context.Context, actionId string) (*platformclientv2.Action, *platformclientv2.APIResponse, error) {
	return p.getIntegrationActionDraftByIdAttr(ctx, p, actionId)
}

func (p *integrationActionsProxy) getIntegrationActionDraftByName(ctx context.Context, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	return p.getIntegrationActionDraftByNameAttr(ctx, p, name)
}

func (p *integrationActionsProxy) updateIntegrationActionDraft(ctx context.Context, actionId string, body platformclientv2.Updatedraftinput) (*platformclientv2.Action, *platformclientv2.APIResponse, error) {
	return p.updateIntegrationActionDraftAttr(ctx, p, actionId, body)
}

func (p *integrationActionsProxy) deleteIntegrationActionDraft(ctx context.Context, actionId string) (*platformclientv2.APIResponse, error) {
	return p.deleteIntegrationActionDraftAttr(ctx, p, actionId)
}

func (p *integrationActionsProxy) getIntegrationActionDraftTemplate(ctx context.Context, actionId string, fileName string) (*string, *platformclientv2.APIResponse, error) {
	return p.getIntegrationActionDraftTemplateAttr(ctx, p, actionId, fileName)
}

func getAllIntegrationActionDraftsFn(ctx context.Context, p *integrationActionsProxy, name string) (*[]platformclientv2.Action, *platformclientv2.APIResponse, error) {
	var allActions []platformclientv2.Action
	var resp *platformclientv2.APIResponse
	var err error
	const pageSize = 100

	actions, resp, err := p.integrationsApi.GetIntegrationsActionsDrafts(pageSize, 1, "", "", "", "", "", name, "", "", "")
	if err != nil {
		return nil, resp, fmt.Errorf("error retrieving actions: %s", err)
	}
	if actions.Entities == nil || len(*actions.Entities) == 0 {
		return &allActions, resp, nil
	}
	allActions = append(allActions, *actions.Entities...)

	for pageNum := 2; pageNum <= *actions.PageCount; pageNum++ {
		actions, resp, err = p.integrationsApi.GetIntegrationsActionsDrafts(pageSize, pageNum, "", "", "", "", "", name, "", "", "")
		if err != nil {
			return nil, resp, fmt.Errorf("error retrieving actions: %s", err)
		}
		if actions.Entities == nil || len(*actions.Entities) == 0 {
			break
		}
		allActions = append(allActions, *actions.Entities...)
	}

	return &allActions, resp, nil
}

func createIntegrationActionDraftFn(ctx context.Context, p *integrationActionsProxy, body platformclientv2.Postactioninput) (*platformclientv2.Action, *platformclientv2.APIResponse, error) {
	return p.integrationsApi.PostIntegrationsActionsDrafts(body)
}

func getIntegrationActionDraftByIdFn(ctx context.Context, p *integrationActionsProxy, actionId string) (*platformclientv2.Action, *platformclientv2.APIResponse, error) {
	return p.integrationsApi.GetIntegrationsActionDraft(actionId, "contract", false, true)
}

func getIntegrationActionDraftByNameFn(ctx context.Context, p *integrationActionsProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	drafts, resp, err := getAllIntegrationActionDraftsFn(ctx, p, name)
	if err != nil {
		return "", false, resp, err
	}

	if drafts == nil || len(*drafts) == 0 {
		return "", true, resp, fmt.Errorf("no integration action draft with name %s", name)
	}

	for _, draft := range *drafts {
		if *draft.Name == name {
			return *draft.Id, false, resp, nil
		}
	}
	return "", true, resp, fmt.Errorf("unable to find integration action draft with name %s", name)
}

func updateIntegrationActionDraftFn(ctx context.Context, p *integrationActionsProxy, actionId string, body platformclientv2.Updatedraftinput) (*platformclientv2.Action, *platformclientv2.APIResponse, error) {
	return p.integrationsApi.PatchIntegrationsActionDraft(actionId, body)
}

func deleteIntegrationActionDraftFn(ctx context.Context, p *integrationActionsProxy, actionId string) (*platformclientv2.APIResponse, error) {
	resp, err := p.integrationsApi.DeleteIntegrationsActionDraft(actionId)
	if err != nil {
		return resp, fmt.Errorf("failed to delete integration action draft: %s", err)
	}
	return resp, nil
}

func getIntegrationActionDraftTemplateFn(ctx context.Context, p *integrationActionsProxy, actionId string, fileName string) (*string, *platformclientv2.APIResponse, error) {
	return p.integrationsApi.GetIntegrationsActionDraftTemplate(actionId, fileName)
}
