package architect_user_prompt

import (
	"context"
	"log"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *architectUserPromptProxy

type createArchitectUserPromptFunc func(ctx context.Context, p *architectUserPromptProxy, body platformclientv2.Prompt) (*platformclientv2.Prompt, *platformclientv2.APIResponse, error)
type getArchitectUserPromptFunc func(ctx context.Context, p *architectUserPromptProxy, id string, includeMediaUris bool, includeResources bool, language []string) (*platformclientv2.Prompt, *platformclientv2.APIResponse, error, bool)
type getAllArchitectUserPromptsFunc func(ctx context.Context, p *architectUserPromptProxy, includeMediaUris bool, includeResources bool, name string) (*[]platformclientv2.Prompt, *platformclientv2.APIResponse, error, bool)
type updateArchitectUserPromptFunc func(ctx context.Context, p *architectUserPromptProxy, id string, body platformclientv2.Prompt) (*platformclientv2.Prompt, *platformclientv2.APIResponse, error)
type deleteArchitectUserPromptFunc func(ctx context.Context, p *architectUserPromptProxy, id string, allResources bool) (*platformclientv2.APIResponse, error)
type createArchitectUserPromptResourceFunc func(ctx context.Context, p *architectUserPromptProxy, id string, body platformclientv2.Promptassetcreate) (*platformclientv2.Promptasset, *platformclientv2.APIResponse, error)
type updateArchitectUserPromptResourceFunc func(ctx context.Context, p *architectUserPromptProxy, id string, languageCode string, body platformclientv2.Promptasset) (*platformclientv2.Promptasset, *platformclientv2.APIResponse, error)
type getArchitectUserPromptIdByNameFunc func(ctx context.Context, p *architectUserPromptProxy, name string) (string, *platformclientv2.APIResponse, error, bool)

// ArchitectUserPromptProxy - proxy for Architect User Prompts
type architectUserPromptProxy struct {
	clientConfig                          *platformclientv2.Configuration
	architectApi                          *platformclientv2.ArchitectApi
	createArchitectUserPromptAttr         createArchitectUserPromptFunc
	getArchitectUserPromptAttr            getArchitectUserPromptFunc
	getAllArchitectUserPromptsAttr        getAllArchitectUserPromptsFunc
	updateArchitectUserPromptAttr         updateArchitectUserPromptFunc
	deleteArchitectUserPromptAttr         deleteArchitectUserPromptFunc
	createArchitectUserPromptResourceAttr createArchitectUserPromptResourceFunc
	updateArchitectUserPromptResourceAttr updateArchitectUserPromptResourceFunc
	getArchitectUserPromptIdByNameAttr    getArchitectUserPromptIdByNameFunc
}

func newArchitectUserPromptProxy(clientConfig *platformclientv2.Configuration) *architectUserPromptProxy {
	api := platformclientv2.NewArchitectApiWithConfig(clientConfig)
	return &architectUserPromptProxy{
		clientConfig:                          clientConfig,
		architectApi:                          api,
		createArchitectUserPromptAttr:         createArchitectUserPromptFn,
		getArchitectUserPromptAttr:            getArchitectUserPromptFn,
		getAllArchitectUserPromptsAttr:        getAllArchitectUserPromptsFn,
		updateArchitectUserPromptAttr:         updateArchitectUserPromptFn,
		deleteArchitectUserPromptAttr:         deleteArchitectUserPromptFn,
		createArchitectUserPromptResourceAttr: createArchitectUserPromptResourceFn,
		updateArchitectUserPromptResourceAttr: updateArchitectUserPromptResourceFn,
		getArchitectUserPromptIdByNameAttr:    getArchitectUserPromptIdByNameFn,
	}
}

func getArchitectUserPromptProxy(clientConfig *platformclientv2.Configuration) *architectUserPromptProxy {
	if internalProxy == nil {
		internalProxy = newArchitectUserPromptProxy(clientConfig)
	}

	return internalProxy
}

// createArchitectUserPrompt creates a new user prompt
func (p *architectUserPromptProxy) createArchitectUserPrompt(ctx context.Context, body platformclientv2.Prompt) (*platformclientv2.Prompt, *platformclientv2.APIResponse, error) {
	return p.createArchitectUserPromptAttr(ctx, p, body)
}

// getArchitectUserPrompt retrieves a user prompt
func (p *architectUserPromptProxy) getArchitectUserPrompt(ctx context.Context, id string, includeMediaUris bool, includeResources bool, language []string) (*platformclientv2.Prompt, *platformclientv2.APIResponse, error, bool) {
	return p.getArchitectUserPromptAttr(ctx, p, id, includeMediaUris, includeResources, language)
}

// getAllArchitectUserPrompts retrieves a list of user prompts
func (p *architectUserPromptProxy) getAllArchitectUserPrompts(ctx context.Context, includeMediaUris bool, includeResources bool, name string) (*[]platformclientv2.Prompt, *platformclientv2.APIResponse, error, bool) {
	return p.getAllArchitectUserPromptsAttr(ctx, p, includeMediaUris, includeResources, name)
}

// updateArchitectUserPrompt updates a user prompt
func (p *architectUserPromptProxy) updateArchitectUserPrompt(ctx context.Context, id string, body platformclientv2.Prompt) (*platformclientv2.Prompt, *platformclientv2.APIResponse, error) {
	return p.updateArchitectUserPromptAttr(ctx, p, id, body)
}

// deleteArchitectUserPrompt deletes a user prompt
func (p *architectUserPromptProxy) deleteArchitectUserPrompt(ctx context.Context, id string, allResources bool) (*platformclientv2.APIResponse, error) {
	return p.deleteArchitectUserPromptAttr(ctx, p, id, allResources)
}

// createArchitectUserPromptResource creates a new user prompt resource
func (p *architectUserPromptProxy) createArchitectUserPromptResource(ctx context.Context, id string, body platformclientv2.Promptassetcreate) (*platformclientv2.Promptasset, *platformclientv2.APIResponse, error) {
	return p.createArchitectUserPromptResourceAttr(ctx, p, id, body)
}

// updateArchitectUserPromptResource updates a user prompt resource
func (p *architectUserPromptProxy) updateArchitectUserPromptResource(ctx context.Context, id string, languageCode string, body platformclientv2.Promptasset) (*platformclientv2.Promptasset, *platformclientv2.APIResponse, error) {
	return p.updateArchitectUserPromptResourceAttr(ctx, p, id, languageCode, body)
}

// getArchitectUserPromptIdByName retrieves a user prompt by name
func (p *architectUserPromptProxy) getArchitectUserPromptIdByName(ctx context.Context, name string) (string, *platformclientv2.APIResponse, error, bool) {
	return p.getArchitectUserPromptIdByNameAttr(ctx, p, name)
}

func createArchitectUserPromptFn(ctx context.Context, p *architectUserPromptProxy, body platformclientv2.Prompt) (*platformclientv2.Prompt, *platformclientv2.APIResponse, error) {
	prompt, response, err := p.architectApi.PostArchitectPrompts(body)
	if err != nil {
		return nil, response, err
	}
	return prompt, response, nil
}

func getArchitectUserPromptFn(ctx context.Context, p *architectUserPromptProxy, id string, includeMediaUris bool, includeResources bool, language []string) (*platformclientv2.Prompt, *platformclientv2.APIResponse, error, bool) {
	prompt, response, err := p.architectApi.GetArchitectPrompt(id, includeMediaUris, includeResources, language)
	if err != nil {
		return nil, response, err, true
	}
	return prompt, response, nil, false
}

func updateArchitectUserPromptFn(ctx context.Context, p *architectUserPromptProxy, id string, body platformclientv2.Prompt) (*platformclientv2.Prompt, *platformclientv2.APIResponse, error) {
	prompt, response, err := p.architectApi.PutArchitectPrompt(id, body)
	if err != nil {
		return nil, response, err
	}
	return prompt, response, nil
}

func deleteArchitectUserPromptFn(ctx context.Context, p *architectUserPromptProxy, id string, allResources bool) (*platformclientv2.APIResponse, error) {
	response, err := p.architectApi.DeleteArchitectPrompt(id, allResources)
	if err != nil {
		return response, err
	}
	return response, nil
}

func getAllArchitectUserPromptsFn(ctx context.Context, p *architectUserPromptProxy, includeMediaUris bool, includeResources bool, name string) (*[]platformclientv2.Prompt, *platformclientv2.APIResponse, error, bool) {
	var (
		pageCount  int
		pageNum    = 1
		allPrompts []platformclientv2.Prompt
		nameString []string
	)

	if len(name) == 0 {
		nameString = nil
	} else {
		nameString = append(nameString, name)
	}

	const pageSize = 100
	userPrompts, response, err := p.architectApi.GetArchitectPrompts(pageNum, pageSize, nameString, "", "", "", "", includeMediaUris, includeResources, nil)

	if err != nil {
		return nil, response, err, true
	}

	if userPrompts != nil && userPrompts.Entities != nil && len(*userPrompts.Entities) > 0 {
		allPrompts = append(allPrompts, *userPrompts.Entities...)
	}
	pageCount = *userPrompts.PageCount

	for pageNum := 2; pageNum <= pageCount; pageNum++ {
		userPrompts, response, getErr := p.architectApi.GetArchitectPrompts(pageNum, pageSize, nameString, "", "", "", "", includeMediaUris, includeResources, nil)
		if getErr != nil {
			return nil, response, getErr, true
		}
		if userPrompts == nil || userPrompts.Entities == nil || len(*userPrompts.Entities) == 0 {
			break
		}
		allPrompts = append(allPrompts, *userPrompts.Entities...)
	}

	return &allPrompts, response, nil, false
}

func createArchitectUserPromptResourceFn(ctx context.Context, p *architectUserPromptProxy, id string, body platformclientv2.Promptassetcreate) (*platformclientv2.Promptasset, *platformclientv2.APIResponse, error) {
	promptAsset, response, err := p.architectApi.PostArchitectPromptResources(id, body)
	if err != nil {
		return nil, response, err
	}
	return promptAsset, response, nil
}

func updateArchitectUserPromptResourceFn(ctx context.Context, p *architectUserPromptProxy, id string, languageCode string, body platformclientv2.Promptasset) (*platformclientv2.Promptasset, *platformclientv2.APIResponse, error) {
	promptAsset, response, err := p.architectApi.PutArchitectPromptResource(id, languageCode, body)
	if err != nil {
		return nil, response, err
	}
	return promptAsset, response, nil
}

func getArchitectUserPromptIdByNameFn(ctx context.Context, p *architectUserPromptProxy, name string) (string, *platformclientv2.APIResponse, error, bool) {
	// Query user prompt by name. Retry in case search has not yet indexed the user prompt.
	prompts, response, err, retryable := p.getAllArchitectUserPrompts(ctx, true, true, name)
	if err != nil {
		return "", response, err, true
	}
	if prompts == nil {
		return "", response, err, true
	}
	for _, prompt := range *prompts {
		if name == *prompt.Name {
			log.Printf("found user prompt id %s by name %s", *prompt.Id, *prompt.Name)
			return *prompt.Id, response, nil, retryable
		}
	}
	return "", response, err, true
}
