package architect_grammar

import (
	"context"
	"fmt"
	"log"

	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"
)

/*
The genesyscloud_architect_grammar_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *architectGrammarProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createArchitectGrammarFunc func(ctx context.Context, p *architectGrammarProxy, grammar *platformclientv2.Grammar) (*platformclientv2.Grammar, error)
type getAllArchitectGrammarFunc func(ctx context.Context, p *architectGrammarProxy) (*[]platformclientv2.Grammar, error)
type getArchitectGrammarByIdFunc func(ctx context.Context, p *architectGrammarProxy, grammarId string) (grammar *platformclientv2.Grammar, responseCode int, err error)
type getArchitectGrammarIdByNameFunc func(ctx context.Context, p *architectGrammarProxy, name string) (grammarId string, retryable bool, err error)
type updateArchitectGrammarFunc func(ctx context.Context, p *architectGrammarProxy, grammarId string, grammar *platformclientv2.Grammar) (*platformclientv2.Grammar, error)
type deleteArchitectGrammarFunc func(ctx context.Context, p *architectGrammarProxy, grammarId string) (responseCode int, err error)

// architectGrammarProxy contains all of the methods that call genesys cloud APIs.
type architectGrammarProxy struct {
	clientConfig                    *platformclientv2.Configuration
	architectApi                    *platformclientv2.ArchitectApi
	createArchitectGrammarAttr      createArchitectGrammarFunc
	getAllArchitectGrammarAttr      getAllArchitectGrammarFunc
	getArchitectGrammarByIdAttr     getArchitectGrammarByIdFunc
	getArchitectGrammarIdByNameAttr getArchitectGrammarIdByNameFunc
	updateArchitectGrammarAttr      updateArchitectGrammarFunc
	deleteArchitectGrammarAttr      deleteArchitectGrammarFunc
}

// newArchitectGrammarProxy initializes the grammar proxy with all of the data needed to communicate with Genesys Cloud
func newArchitectGrammarProxy(clientConfig *platformclientv2.Configuration) *architectGrammarProxy {
	api := platformclientv2.NewArchitectApiWithConfig(clientConfig)
	return &architectGrammarProxy{
		clientConfig:                    clientConfig,
		architectApi:                    api,
		createArchitectGrammarAttr:      createArchitectGrammarFn,
		getAllArchitectGrammarAttr:      getAllArchitectGrammarFn,
		getArchitectGrammarByIdAttr:     getArchitectGrammarByIdFn,
		getArchitectGrammarIdByNameAttr: getArchitectGrammarIdByNameFn,
		updateArchitectGrammarAttr:      updateArchitectGrammarFn,
		deleteArchitectGrammarAttr:      deleteArchitectGrammarFn,
	}
}

// getArchitectGrammarProxy acts as a singleton for the internalProxy. It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getArchitectGrammarProxy(clientConfig *platformclientv2.Configuration) *architectGrammarProxy {
	if internalProxy == nil {
		internalProxy = newArchitectGrammarProxy(clientConfig)
	}

	return internalProxy
}

// createArchitectGrammar creates a Genesys Cloud Architect Grammar
func (p *architectGrammarProxy) createArchitectGrammar(ctx context.Context, grammar *platformclientv2.Grammar) (*platformclientv2.Grammar, error) {
	return p.createArchitectGrammarAttr(ctx, p, grammar)
}

// getAllArchitectGrammar retrieves all Genesys Cloud Architect Grammar
func (p *architectGrammarProxy) getAllArchitectGrammar(ctx context.Context) (*[]platformclientv2.Grammar, error) {
	return p.getAllArchitectGrammarAttr(ctx, p)
}

// getArchitectGrammarById returns a single Genesys Cloud Architect Grammar by Id
func (p *architectGrammarProxy) getArchitectGrammarById(ctx context.Context, grammarId string) (grammar *platformclientv2.Grammar, statusCode int, err error) {
	return p.getArchitectGrammarByIdAttr(ctx, p, grammarId)
}

// getArchitectGrammarIdByName returns a single Genesys Cloud Architect Grammar by a name
func (p *architectGrammarProxy) getArchitectGrammarIdByName(ctx context.Context, name string) (grammarId string, retryable bool, err error) {
	return p.getArchitectGrammarIdByNameAttr(ctx, p, name)
}

// updateArchitectGrammar updates a Genesys Cloud Architect Grammar
func (p *architectGrammarProxy) updateArchitectGrammar(ctx context.Context, grammarId string, grammar *platformclientv2.Grammar) (*platformclientv2.Grammar, error) {
	return p.updateArchitectGrammarAttr(ctx, p, grammarId, grammar)
}

// deleteArchitectGrammar deletes a Genesys Cloud Architect Grammar by Id
func (p *architectGrammarProxy) deleteArchitectGrammar(ctx context.Context, grammarId string) (statusCode int, err error) {
	return p.deleteArchitectGrammarAttr(ctx, p, grammarId)
}

// createArchitectGrammarFn is an implementation function for creating a Genesys Cloud Architect Grammar
func createArchitectGrammarFn(ctx context.Context, p *architectGrammarProxy, grammar *platformclientv2.Grammar) (*platformclientv2.Grammar, error) {
	grammarSdk, _, err := p.architectApi.PostArchitectGrammars(*grammar)
	if err != nil {
		return nil, fmt.Errorf("Failed to create grammar: %s", err)
	}

	return grammarSdk, nil
}

// getAllArchitectGrammarFn is the implementation for retrieving all Architect Grammars in Genesys Cloud
func getAllArchitectGrammarFn(ctx context.Context, p *architectGrammarProxy) (*[]platformclientv2.Grammar, error) {
	var allGrammars []platformclientv2.Grammar

	grammars, _, err := p.architectApi.GetArchitectGrammars(1, 100, "", "", []string{}, "", "", "", true)
	if err != nil {
		return nil, fmt.Errorf("Failed to get architect grammars: %v", err)
	}
	if grammars.Entities == nil || len(*grammars.Entities) == 0 {
		return &allGrammars, nil
	}
	for _, grammar := range *grammars.Entities {
		allGrammars = append(allGrammars, grammar)
	}

	for pageNum := 2; pageNum <= *grammars.PageCount; pageNum++ {
		const pageSize = 100

		grammars, _, err := p.architectApi.GetArchitectGrammars(pageNum, pageSize, "", "", []string{}, "", "", "", true)
		if err != nil {
			return nil, fmt.Errorf("Failed to get architect grammars: %v", err)
		}

		if grammars.Entities == nil || len(*grammars.Entities) == 0 {
			break
		}

		for _, grammar := range *grammars.Entities {
			allGrammars = append(allGrammars, grammar)
		}
	}

	return &allGrammars, nil
}

// getArchitectGrammarByIdFn is an implementation of the function to get a Genesys Cloud Architect Grammar by Id
func getArchitectGrammarByIdFn(ctx context.Context, p *architectGrammarProxy, grammarId string) (grammar *platformclientv2.Grammar, statusCode int, err error) {
	grammar, resp, err := p.architectApi.GetArchitectGrammar(grammarId, true)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("Failed to retrieve grammar by id %s: %s", grammarId, err)
	}
	return grammar, resp.StatusCode, nil
}

// getArchitectGrammarIdByNameFn is an implementation of the function to get a Genesys Cloud Architect Grammar by name
func getArchitectGrammarIdByNameFn(ctx context.Context, p *architectGrammarProxy, name string) (grammarId string, retryable bool, err error) {
	grammars, err := getAllArchitectGrammarFn(ctx, p)
	if err != nil {
		return "", false, err
	}
	if grammars == nil || len(*grammars) == 0 {
		return "", true, fmt.Errorf("No architect grammars found with name %s", name)
	}

	var grammar platformclientv2.Grammar
	for _, grammarSdk := range *grammars {
		if *grammarSdk.Name == name {
			log.Printf("Retrieved the grammar id %s by name %s", *grammarSdk.Id, name)
			grammar = grammarSdk
			return *grammar.Id, false, nil
		}
	}

	return "", false, fmt.Errorf("Unable to find grammar with name %s", name)
}

// updateArchitectGrammarFn is an implementation of the function to update a Genesys Cloud Architect Grammar
func updateArchitectGrammarFn(ctx context.Context, p *architectGrammarProxy, grammarId string, grammar *platformclientv2.Grammar) (*platformclientv2.Grammar, error) {
	grammarSdk, _, err := p.architectApi.PatchArchitectGrammar(grammarId, *grammar)
	if err != nil {
		return nil, fmt.Errorf("Failed to update grammar %s: %s", grammarId, err)
	}

	return grammarSdk, nil
}

// deleteArchitectGrammarFn is an implementation function for deleting a Genesys Cloud Architect Grammar
func deleteArchitectGrammarFn(ctx context.Context, p *architectGrammarProxy, grammarId string) (statusCode int, err error) {
	_, resp, err := p.architectApi.DeleteArchitectGrammar(grammarId)
	if err != nil {
		return resp.StatusCode, fmt.Errorf("Failed to delete grammar: %s", err)
	}

	return resp.StatusCode, nil
}
