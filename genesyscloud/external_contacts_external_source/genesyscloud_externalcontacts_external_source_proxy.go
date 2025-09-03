package external_contacts_external_source

import (
	"context"
	"fmt"
	"log"

	rc "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *externalContactsExternalSourceProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createExternalContactsExternalSourceFunc func(ctx context.Context, p *externalContactsExternalSourceProxy, externalSource *platformclientv2.Externalsource) (*platformclientv2.Externalsource, *platformclientv2.APIResponse, error)
type getAllExternalContactsExternalSourceFunc func(ctx context.Context, p *externalContactsExternalSourceProxy, query string) (*[]platformclientv2.Externalsource, *platformclientv2.APIResponse, error)
type getExternalContactsExternalSourceIdByNameFunc func(ctx context.Context, p *externalContactsExternalSourceProxy, name string) (id string, retryable bool, apiResponse *platformclientv2.APIResponse, err error)
type getExternalContactsExternalSourceByIdFunc func(ctx context.Context, p *externalContactsExternalSourceProxy, id string) (externalSource *platformclientv2.Externalsource, apiResponse *platformclientv2.APIResponse, err error)
type updateExternalContactsExternalSourceFunc func(ctx context.Context, p *externalContactsExternalSourceProxy, id string, externalSource *platformclientv2.Externalsource) (*platformclientv2.Externalsource, *platformclientv2.APIResponse, error)
type deleteExternalContactsExternalSourceFunc func(ctx context.Context, p *externalContactsExternalSourceProxy, id string) (apiResponse *platformclientv2.APIResponse, err error)

// externalContactsExternalSourceProxy contains all of the methods that call genesys cloud APIs.
type externalContactsExternalSourceProxy struct {
	clientConfig                                  *platformclientv2.Configuration
	externalContactsApi                           *platformclientv2.ExternalContactsApi
	createExternalContactsExternalSourceAttr      createExternalContactsExternalSourceFunc
	getAllExternalContactsExternalSourceAttr      getAllExternalContactsExternalSourceFunc
	getExternalContactsExternalSourceIdByNameAttr getExternalContactsExternalSourceIdByNameFunc
	getExternalContactsExternalSourceByIdAttr     getExternalContactsExternalSourceByIdFunc
	updateExternalContactsExternalSourceAttr      updateExternalContactsExternalSourceFunc
	deleteExternalContactsExternalSourceAttr      deleteExternalContactsExternalSourceFunc
	externalSourcesCache                          rc.CacheInterface[platformclientv2.Externalsource]
}

// newExternalContactsExternalSourceProxy initializes the external contacts external source proxy with all of the data needed to communicate with Genesys Cloud
func newExternalContactsExternalSourceProxy(clientConfig *platformclientv2.Configuration) *externalContactsExternalSourceProxy {
	api := platformclientv2.NewExternalContactsApiWithConfig(clientConfig)
	externalSourcesCache := rc.NewResourceCache[platformclientv2.Externalsource]()
	return &externalContactsExternalSourceProxy{
		clientConfig:                                  clientConfig,
		externalContactsApi:                           api,
		createExternalContactsExternalSourceAttr:      createExternalContactsExternalSourceFn,
		getAllExternalContactsExternalSourceAttr:      getAllExternalContactsExternalSourceFn,
		getExternalContactsExternalSourceIdByNameAttr: getExternalContactsExternalSourceIdByNameFn,
		getExternalContactsExternalSourceByIdAttr:     getExternalContactsExternalSourceByIdFn,
		updateExternalContactsExternalSourceAttr:      updateExternalContactsExternalSourceFn,
		deleteExternalContactsExternalSourceAttr:      deleteExternalContactsExternalSourceFn,
		externalSourcesCache:                          externalSourcesCache,
	}
}

// getExternalContactsExternalSourceProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getExternalContactsExternalSourceProxy(clientConfig *platformclientv2.Configuration) *externalContactsExternalSourceProxy {
	if internalProxy == nil {
		internalProxy = newExternalContactsExternalSourceProxy(clientConfig)
	}

	return internalProxy
}

// createExternalContactsExternalSource creates a Genesys Cloud external contacts external source
func (p *externalContactsExternalSourceProxy) createExternalContactsExternalSource(ctx context.Context, externalContactsExternalSource *platformclientv2.Externalsource) (*platformclientv2.Externalsource, *platformclientv2.APIResponse, error) {
	return p.createExternalContactsExternalSourceAttr(ctx, p, externalContactsExternalSource)
}

// getExternalContactsExternalSource retrieves all Genesys Cloud external contacts external source
func (p *externalContactsExternalSourceProxy) getAllExternalContactsExternalSources(ctx context.Context, name string) (*[]platformclientv2.Externalsource, *platformclientv2.APIResponse, error) {
	return p.getAllExternalContactsExternalSourceAttr(ctx, p, name)
}

// getExternalContactsExternalSourceIdByName returns a single Genesys Cloud external contacts external source by a name
func (p *externalContactsExternalSourceProxy) getExternalContactsExternalSourceIdByName(ctx context.Context, name string) (id string, retryable bool, apiResponse *platformclientv2.APIResponse, err error) {
	return p.getExternalContactsExternalSourceIdByNameAttr(ctx, p, name)
}

// getExternalContactsExternalSourceById returns a single Genesys Cloud external contacts external source by Id
func (p *externalContactsExternalSourceProxy) getExternalContactsExternalSourceById(ctx context.Context, id string) (externalContactsExternalSource *platformclientv2.Externalsource, apiResponse *platformclientv2.APIResponse, err error) {
	return p.getExternalContactsExternalSourceByIdAttr(ctx, p, id)
}

// updateExternalContactsExternalSource updates a Genesys Cloud external contacts external source
func (p *externalContactsExternalSourceProxy) updateExternalContactsExternalSource(ctx context.Context, id string, externalContactsExternalSource *platformclientv2.Externalsource) (*platformclientv2.Externalsource, *platformclientv2.APIResponse, error) {
	return p.updateExternalContactsExternalSourceAttr(ctx, p, id, externalContactsExternalSource)
}

// deleteExternalContactsExternalSource deletes a Genesys Cloud external contacts external source by Id
func (p *externalContactsExternalSourceProxy) deleteExternalContactsExternalSource(ctx context.Context, id string) (apiResponse *platformclientv2.APIResponse, err error) {
	return p.deleteExternalContactsExternalSourceAttr(ctx, p, id)
}

// createExternalContactsExternalSourceFn is an implementation function for creating a Genesys Cloud external contacts external source
func createExternalContactsExternalSourceFn(ctx context.Context, p *externalContactsExternalSourceProxy, externalContactsExternalSource *platformclientv2.Externalsource) (*platformclientv2.Externalsource, *platformclientv2.APIResponse, error) {
	return p.externalContactsApi.PostExternalcontactsExternalsources(*externalContactsExternalSource)
}

// getAllExternalContactsExternalSourceFn is the implementation for retrieving all external contacts external source in Genesys Cloud
func getAllExternalContactsExternalSourceFn(ctx context.Context, p *externalContactsExternalSourceProxy, query string) (*[]platformclientv2.Externalsource, *platformclientv2.APIResponse, error) {
	var allExternalSources []platformclientv2.Externalsource
	var response *platformclientv2.APIResponse

	cursor := ""
	for {
		externalSources, resp, err := p.externalContactsApi.GetExternalcontactsExternalsources(cursor, 100, query, true) // workaround for active being a required arg on Platform Client SDK Go GetExternalcontactsExternalsources
		if err != nil {
			return nil, resp, fmt.Errorf("failed to get external sources: %v", err)
		}
		response = resp
		if externalSources.Entities == nil || len(*externalSources.Entities) == 0 {
			break
		}

		allExternalSources = append(allExternalSources, *externalSources.Entities...)

		if externalSources.Cursors == nil || externalSources.Cursors.After == nil {
			break
		}
		cursor = *externalSources.Cursors.After
	}

	cursor = ""
	for {
		externalSources, resp, err := p.externalContactsApi.GetExternalcontactsExternalsources(cursor, 100, query, false) // workaround for active being a required arg on Platform Client SDK Go GetExternalcontactsExternalsources
		if err != nil {
			return nil, resp, fmt.Errorf("failed to get external sources: %v", err)
		}
		response = resp
		if externalSources.Entities == nil || len(*externalSources.Entities) == 0 {
			break
		}

		allExternalSources = append(allExternalSources, *externalSources.Entities...)

		if externalSources.Cursors == nil || externalSources.Cursors.After == nil {
			break
		}
		cursor = *externalSources.Cursors.After
	}

	// Cache the External Contacts resource into the p.externalContactsCache for later use
	for _, externalSource := range allExternalSources {
		if externalSource.Id == nil {
			continue
		}
		rc.SetCache(p.externalSourcesCache, *externalSource.Id, externalSource)
	}

	return &allExternalSources, response, nil
}

// getExternalContactsExternalSourceIdByNameFn is an implementation of the function to get a Genesys Cloud external contacts external source by name
func getExternalContactsExternalSourceIdByNameFn(ctx context.Context, p *externalContactsExternalSourceProxy, name string) (id string, retryable bool, apiResponse *platformclientv2.APIResponse, err error) {

	externalSources, response, err := p.getAllExternalContactsExternalSources(ctx, name)
	if err != nil {
		return "", false, response, err
	}

	if externalSources == nil || len(*externalSources) == 0 {
		return "", true, response, fmt.Errorf("no external sources found with name %s", name)
	}

	var externalSource platformclientv2.Externalsource
	for _, externalSourceSdk := range *externalSources {
		if *externalSourceSdk.Name == name {
			log.Printf("Retrieved the external source id %s by name %s", *externalSourceSdk.Id, name)
			externalSource = externalSourceSdk
			return *externalSource.Id, false, response, nil
		}
	}

	return "", true, response, fmt.Errorf("unable to find external sources with name %s", name)
}

// getExternalContactsExternalSourceByIdFn is an implementation of the function to get a Genesys Cloud external contacts external source by Id
func getExternalContactsExternalSourceByIdFn(ctx context.Context, p *externalContactsExternalSourceProxy, id string) (externalContactsExternalSource *platformclientv2.Externalsource, apiResponse *platformclientv2.APIResponse, err error) {
	if externalSource := rc.GetCacheItem(p.externalSourcesCache, id); externalSource != nil {
		return externalSource, nil, nil
	}
	return p.externalContactsApi.GetExternalcontactsExternalsource(id)
}

// updateExternalContactsExternalSourceFn is an implementation of the function to update a Genesys Cloud external contacts external source
func updateExternalContactsExternalSourceFn(ctx context.Context, p *externalContactsExternalSourceProxy, id string, externalContactsExternalSource *platformclientv2.Externalsource) (*platformclientv2.Externalsource, *platformclientv2.APIResponse, error) {
	return p.externalContactsApi.PutExternalcontactsExternalsource(id, *externalContactsExternalSource)
}

// deleteExternalContactsExternalSourceFn is an implementation function for deleting a Genesys Cloud external contacts external source
func deleteExternalContactsExternalSourceFn(ctx context.Context, p *externalContactsExternalSourceProxy, id string) (apiResponse *platformclientv2.APIResponse, err error) {
	_, response, err := p.externalContactsApi.DeleteExternalcontactsExternalsource(id)
	if err != nil {
		return response, err
	}
	rc.DeleteCacheItem(p.externalSourcesCache, id)
	return response, err
}
