package architect_grammar

import (
	"context"
	"fmt"
	"log"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The genesyscloud_architect_grammar_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *architectGrammarProxy

// Type definitions for each func on our proxy so that we can easily mock them out later
type createArchitectGrammarFunc func(ctx context.Context, p *architectGrammarProxy, grammar *platformclientv2.Grammar) (*platformclientv2.Grammar, *platformclientv2.APIResponse, error)
type getAllArchitectGrammarFunc func(ctx context.Context, p *architectGrammarProxy) (*[]platformclientv2.Grammar, *platformclientv2.APIResponse, error)
type getArchitectGrammarByIdFunc func(ctx context.Context, p *architectGrammarProxy, grammarId string) (*platformclientv2.Grammar, *platformclientv2.APIResponse, error)
type getArchitectGrammarIdByNameFunc func(ctx context.Context, p *architectGrammarProxy, name string) (string, bool, *platformclientv2.APIResponse, error)
type updateArchitectGrammarFunc func(ctx context.Context, p *architectGrammarProxy, grammarId string, grammar *platformclientv2.Grammar) (*platformclientv2.Grammar, *platformclientv2.APIResponse, error)
type deleteArchitectGrammarFunc func(ctx context.Context, p *architectGrammarProxy, grammarId string) (*platformclientv2.APIResponse, error)

// architectGrammarProxy contains all the methods that call genesys cloud APIs.
type architectGrammarProxy struct {
	clientConfig                    *platformclientv2.Configuration
	architectApi                    *platformclientv2.ArchitectApi
	createArchitectGrammarAttr      createArchitectGrammarFunc
	getAllArchitectGrammarAttr      getAllArchitectGrammarFunc
	getArchitectGrammarByIdAttr     getArchitectGrammarByIdFunc
	getArchitectGrammarIdByNameAttr getArchitectGrammarIdByNameFunc
	updateArchitectGrammarAttr      updateArchitectGrammarFunc
	deleteArchitectGrammarAttr      deleteArchitectGrammarFunc
	grammarCache                    rc.CacheInterface[platformclientv2.Grammar]
}

// newArchitectGrammarProxy initializes the grammar proxy with all the data needed to communicate with Genesys Cloud
func newArchitectGrammarProxy(clientConfig *platformclientv2.Configuration) *architectGrammarProxy {
	api := platformclientv2.NewArchitectApiWithConfig(clientConfig)
	grammarCache := rc.NewResourceCache[platformclientv2.Grammar]()
	return &architectGrammarProxy{
		clientConfig:                    clientConfig,
		architectApi:                    api,
		createArchitectGrammarAttr:      createArchitectGrammarFn,
		getAllArchitectGrammarAttr:      getAllArchitectGrammarFn,
		getArchitectGrammarByIdAttr:     getArchitectGrammarByIdFn,
		getArchitectGrammarIdByNameAttr: getArchitectGrammarIdByNameFn,
		updateArchitectGrammarAttr:      updateArchitectGrammarFn,
		deleteArchitectGrammarAttr:      deleteArchitectGrammarFn,
		grammarCache:                    grammarCache,
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
func (p *architectGrammarProxy) createArchitectGrammar(ctx context.Context, grammar *platformclientv2.Grammar) (*platformclientv2.Grammar, *platformclientv2.APIResponse, error) {
	return p.createArchitectGrammarAttr(ctx, p, grammar)
}

// getAllArchitectGrammar retrieves all Genesys Cloud Architect Grammar
func (p *architectGrammarProxy) getAllArchitectGrammar(ctx context.Context) (*[]platformclientv2.Grammar, *platformclientv2.APIResponse, error) {
	return p.getAllArchitectGrammarAttr(ctx, p)
}

// getArchitectGrammarById returns a single Genesys Cloud Architect Grammar by ID
func (p *architectGrammarProxy) getArchitectGrammarById(ctx context.Context, grammarId string) (*platformclientv2.Grammar, *platformclientv2.APIResponse, error) {
	return p.getArchitectGrammarByIdAttr(ctx, p, grammarId)
}

// getArchitectGrammarIdByName returns a single Genesys Cloud Architect Grammar by a name
func (p *architectGrammarProxy) getArchitectGrammarIdByName(ctx context.Context, name string) (string, bool, *platformclientv2.APIResponse, error) {
	return p.getArchitectGrammarIdByNameAttr(ctx, p, name)
}

// updateArchitectGrammar updates a Genesys Cloud Architect Grammar
func (p *architectGrammarProxy) updateArchitectGrammar(ctx context.Context, grammarId string, grammar *platformclientv2.Grammar) (*platformclientv2.Grammar, *platformclientv2.APIResponse, error) {
	return p.updateArchitectGrammarAttr(ctx, p, grammarId, grammar)
}

// deleteArchitectGrammar deletes a Genesys Cloud Architect Grammar by ID
func (p *architectGrammarProxy) deleteArchitectGrammar(ctx context.Context, grammarId string) (*platformclientv2.APIResponse, error) {
	return p.deleteArchitectGrammarAttr(ctx, p, grammarId)
}

// createArchitectGrammarFn is an implementation function for creating a Genesys Cloud Architect Grammar
func createArchitectGrammarFn(_ context.Context, p *architectGrammarProxy, grammar *platformclientv2.Grammar) (*platformclientv2.Grammar, *platformclientv2.APIResponse, error) {
	grammarSdk, resp, err := p.architectApi.PostArchitectGrammars(*grammar)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to create grammar: %s %v", err, resp)
	}
	return grammarSdk, resp, nil
}

// getAllArchitectGrammarFn is the implementation for retrieving all Architect Grammars in Genesys Cloud
func getAllArchitectGrammarFn(_ context.Context, p *architectGrammarProxy) (*[]platformclientv2.Grammar, *platformclientv2.APIResponse, error) {
	var allGrammars []platformclientv2.Grammar

	grammars, resp, err := p.architectApi.GetArchitectGrammars(1, 100, "", "", []string{}, "", "", "", true)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to get architect grammars: %v %v", err, resp)
	}
	if grammars.Entities == nil || len(*grammars.Entities) == 0 {
		return &allGrammars, resp, nil
	}
	allGrammars = append(allGrammars, *grammars.Entities...)

	for pageNum := 2; pageNum <= *grammars.PageCount; pageNum++ {
		const pageSize = 100

		grammars, resp, err := p.architectApi.GetArchitectGrammars(pageNum, pageSize, "", "", []string{}, "", "", "", true)
		if err != nil {
			return nil, resp, fmt.Errorf("failed to get architect grammars: %v %v", err, resp)
		}

		if grammars.Entities == nil || len(*grammars.Entities) == 0 {
			break
		}

		allGrammars = append(allGrammars, *grammars.Entities...)
	}

	for _, grammar := range allGrammars {
		rc.SetCache(p.grammarCache, *grammar.Id, grammar)
	}

	return &allGrammars, resp, nil
}

// getArchitectGrammarByIdFn is an implementation of the function to get a Genesys Cloud Architect Grammar by ID
func getArchitectGrammarByIdFn(_ context.Context, p *architectGrammarProxy, grammarId string) (*platformclientv2.Grammar, *platformclientv2.APIResponse, error) {
	grammar := rc.GetCacheItem(p.grammarCache, grammarId)
	if grammar != nil {
		return grammar, nil, nil
	}

	grammar, resp, err := p.architectApi.GetArchitectGrammar(grammarId, true)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to retrieve grammar by id %s: %s", grammarId, err)
	}
	return grammar, resp, nil
}

// getArchitectGrammarIdByNameFn is an implementation of the function to get a Genesys Cloud Architect Grammar by name
func getArchitectGrammarIdByNameFn(ctx context.Context, p *architectGrammarProxy, name string) (string, bool, *platformclientv2.APIResponse, error) {
	grammars, resp, err := getAllArchitectGrammarFn(ctx, p)
	if err != nil {
		return "", false, resp, err
	}
	if grammars == nil || len(*grammars) == 0 {
		return "", true, resp, fmt.Errorf("no architect grammars found with name %s", name)
	}

	var grammar platformclientv2.Grammar
	for _, grammarSdk := range *grammars {
		if *grammarSdk.Name == name {
			log.Printf("Retrieved the grammar id %s by name %s", *grammarSdk.Id, name)
			grammar = grammarSdk
			return *grammar.Id, false, resp, nil
		}
	}

	return "", false, resp, fmt.Errorf("unable to find grammar with name %s", name)
}

// updateArchitectGrammarFn is an implementation of the function to update a Genesys Cloud Architect Grammar
func updateArchitectGrammarFn(_ context.Context, p *architectGrammarProxy, grammarId string, grammar *platformclientv2.Grammar) (*platformclientv2.Grammar, *platformclientv2.APIResponse, error) {
	grammarSdk, resp, err := p.architectApi.PatchArchitectGrammar(grammarId, *grammar)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to update grammar %s: %s", grammarId, err)
	}
	return grammarSdk, resp, nil
}

// deleteArchitectGrammarFn is an implementation function for deleting a Genesys Cloud Architect Grammar
func deleteArchitectGrammarFn(_ context.Context, p *architectGrammarProxy, grammarId string) (*platformclientv2.APIResponse, error) {
	_, resp, err := p.architectApi.DeleteArchitectGrammar(grammarId)
	if err != nil {
		return resp, fmt.Errorf("failed to delete grammar: %s", err)
	}
	return resp, nil
}
