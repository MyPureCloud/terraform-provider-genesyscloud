package routing_language

import (
	"context"
	"fmt"
	"log"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

var internalProxy *routingLanguageProxy

type getAllRoutingLanguagesFunc func(ctx context.Context, p *routingLanguageProxy, name string) (*[]platformclientv2.Language, *platformclientv2.APIResponse, error)
type createRoutingLanguageFunc func(ctx context.Context, p *routingLanguageProxy, language *platformclientv2.Language) (*platformclientv2.Language, *platformclientv2.APIResponse, error)
type getRoutingLanguageByIdFunc func(ctx context.Context, p *routingLanguageProxy, id string) (*platformclientv2.Language, *platformclientv2.APIResponse, error)
type getRoutingLanguageIdByNameFunc func(ctx context.Context, p *routingLanguageProxy, name string) (string, *platformclientv2.APIResponse, bool, error)
type deleteRoutingLanguageFunc func(ctx context.Context, p *routingLanguageProxy, id string) (*platformclientv2.APIResponse, error)

// routingLanguageProxy contains all of the methods that call genesys cloud APIs.
type routingLanguageProxy struct {
	clientConfig                   *platformclientv2.Configuration
	routingApi                     *platformclientv2.RoutingApi
	createRoutingLanguageAttr      createRoutingLanguageFunc
	getAllRoutingLanguagesAttr     getAllRoutingLanguagesFunc
	getRoutingLanguageIdByNameAttr getRoutingLanguageIdByNameFunc
	getRoutingLanguageByIdAttr     getRoutingLanguageByIdFunc
	deleteRoutingLanguageAttr      deleteRoutingLanguageFunc
	routingLanguageCache           rc.CacheInterface[platformclientv2.Language]
}

// newRoutingLanguageProxy initializes the routing language proxy with all of the data needed to communicate with Genesys Cloud
func newRoutingLanguageProxy(clientConfig *platformclientv2.Configuration) *routingLanguageProxy {
	api := platformclientv2.NewRoutingApiWithConfig(clientConfig)
	routingLanguageCache := rc.NewResourceCache[platformclientv2.Language]()
	return &routingLanguageProxy{
		clientConfig:                   clientConfig,
		routingApi:                     api,
		createRoutingLanguageAttr:      createRoutingLanguageFn,
		getAllRoutingLanguagesAttr:     getAllRoutingLanguagesFn,
		getRoutingLanguageIdByNameAttr: getRoutingLanguageIdByNameFn,
		getRoutingLanguageByIdAttr:     getRoutingLanguageByIdFn,
		deleteRoutingLanguageAttr:      deleteRoutingLanguageFn,
		routingLanguageCache:           routingLanguageCache,
	}
}

func getRoutingLanguageProxy(clientConfig *platformclientv2.Configuration) *routingLanguageProxy {
	if internalProxy == nil {
		internalProxy = newRoutingLanguageProxy(clientConfig)
	}
	return internalProxy
}

// getRoutingLanguage retrieves all Genesys Cloud routing language
func (p *routingLanguageProxy) getAllRoutingLanguages(ctx context.Context, name string) (*[]platformclientv2.Language, *platformclientv2.APIResponse, error) {
	return p.getAllRoutingLanguagesAttr(ctx, p, name)
}

// createRoutingLanguage creates a Genesys Cloud routing language
func (p *routingLanguageProxy) createRoutingLanguage(ctx context.Context, routingLanguage *platformclientv2.Language) (*platformclientv2.Language, *platformclientv2.APIResponse, error) {
	return p.createRoutingLanguageAttr(ctx, p, routingLanguage)
}

// getRoutingLanguageById returns a single Genesys Cloud routing language by Id
func (p *routingLanguageProxy) getRoutingLanguageById(ctx context.Context, id string) (*platformclientv2.Language, *platformclientv2.APIResponse, error) {
	return p.getRoutingLanguageByIdAttr(ctx, p, id)
}

// getRoutingLanguageIdByName returns a single Genesys Cloud routing language by a name
func (p *routingLanguageProxy) getRoutingLanguageIdByName(ctx context.Context, name string) (string, *platformclientv2.APIResponse, bool, error) {
	return p.getRoutingLanguageIdByNameAttr(ctx, p, name)
}

// deleteRoutingLanguage deletes a Genesys Cloud routing language by Id
func (p *routingLanguageProxy) deleteRoutingLanguage(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	return p.deleteRoutingLanguageAttr(ctx, p, id)
}

// getAllRoutingLanguageFn is the implementation for retrieving all routing language in Genesys Cloud
func getAllRoutingLanguagesFn(ctx context.Context, p *routingLanguageProxy, name string) (*[]platformclientv2.Language, *platformclientv2.APIResponse, error) {
	var (
		allLanguages []platformclientv2.Language
		response     *platformclientv2.APIResponse
		pageSize     = 100
	)

	languages, resp, err := p.routingApi.GetRoutingLanguages(pageSize, 1, "", name, []string{})
	if err != nil {
		return nil, resp, fmt.Errorf("failed to get language: %v", err)
	}

	if languages.Entities == nil || len(*languages.Entities) == 0 {
		return &allLanguages, resp, nil
	}
	allLanguages = append(allLanguages, *languages.Entities...)

	for pageNum := 2; pageNum <= *languages.PageCount; pageNum++ {
		languages, resp, err := p.routingApi.GetRoutingLanguages(pageSize, pageNum, "", name, []string{})
		if err != nil {
			return nil, resp, fmt.Errorf("failed to get language: %v", err)
		}

		response = resp
		if languages.Entities == nil || len(*languages.Entities) == 0 {
			break
		}
		allLanguages = append(allLanguages, *languages.Entities...)
	}

	for _, language := range allLanguages {
		rc.SetCache(p.routingLanguageCache, *language.Id, language)
	}

	return &allLanguages, response, nil
}

// createRoutingLanguageFn is an implementation function for creating a Genesys Cloud routing language
func createRoutingLanguageFn(ctx context.Context, p *routingLanguageProxy, routingLanguage *platformclientv2.Language) (*platformclientv2.Language, *platformclientv2.APIResponse, error) {
	return p.routingApi.PostRoutingLanguages(*routingLanguage)
}

// getRoutingLanguageByIdFn is an implementation of the function to get a Genesys Cloud routing language by Id
func getRoutingLanguageByIdFn(ctx context.Context, p *routingLanguageProxy, id string) (*platformclientv2.Language, *platformclientv2.APIResponse, error) {
	if language := rc.GetCacheItem(p.routingLanguageCache, id); language != nil {
		return language, nil, nil
	}
	return p.routingApi.GetRoutingLanguage(id)
}

// getRoutingLanguageIdByNameFn is an implementation of the function to get a Genesys Cloud routing language by name
func getRoutingLanguageIdByNameFn(ctx context.Context, p *routingLanguageProxy, name string) (string, *platformclientv2.APIResponse, bool, error) {
	languages, resp, err := getAllRoutingLanguagesFn(ctx, p, name)
	if err != nil {
		return "", resp, false, err
	}

	if languages == nil || len(*languages) == 0 {
		return "", resp, true, fmt.Errorf("no routing language found with name %s", name)
	}

	for _, language := range *languages {
		if *language.Name == name {
			log.Printf("Retrieved the routing language id %s by name %s", *language.Id, name)
			return *language.Id, resp, false, nil
		}
	}
	return "", resp, true, fmt.Errorf("unable to find routing language with name %s", name)
}

// deleteRoutingLanguageFn is an implementation function for deleting a Genesys Cloud routing language
func deleteRoutingLanguageFn(ctx context.Context, p *routingLanguageProxy, id string) (*platformclientv2.APIResponse, error) {
	return p.routingApi.DeleteRoutingLanguage(id)
}
