package external_contacts_organization

import (
	"context"
	"fmt"
	"log"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v149/platformclientv2"
)

/*
The genesyscloud_external_contacts_organization_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *externalContactsOrganizationProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createExternalContactsOrganizationFunc func(ctx context.Context, p *externalContactsOrganizationProxy, externalOrganization *platformclientv2.Externalorganization) (*platformclientv2.Externalorganization, *platformclientv2.APIResponse, error)
type getAllExternalContactsOrganizationFunc func(ctx context.Context, p *externalContactsOrganizationProxy, query string) (*[]platformclientv2.Externalorganization, *platformclientv2.APIResponse, error)
type getExternalContactsOrganizationIdByNameFunc func(ctx context.Context, p *externalContactsOrganizationProxy, name string) (id string, retryable bool, apiResponse *platformclientv2.APIResponse, err error)
type getExternalContactsOrganizationByIdFunc func(ctx context.Context, p *externalContactsOrganizationProxy, id string) (externalOrganization *platformclientv2.Externalorganization, apiResponse *platformclientv2.APIResponse, err error)
type updateExternalContactsOrganizationFunc func(ctx context.Context, p *externalContactsOrganizationProxy, id string, externalOrganization *platformclientv2.Externalorganization) (*platformclientv2.Externalorganization, *platformclientv2.APIResponse, error)
type deleteExternalContactsOrganizationFunc func(ctx context.Context, p *externalContactsOrganizationProxy, id string) (apiResponse *platformclientv2.APIResponse, err error)

// externalContactsOrganizationProxy contains all of the methods that call genesys cloud APIs.
type externalContactsOrganizationProxy struct {
	clientConfig                                *platformclientv2.Configuration
	externalContactsApi                         *platformclientv2.ExternalContactsApi
	createExternalContactsOrganizationAttr      createExternalContactsOrganizationFunc
	getAllExternalContactsOrganizationAttr      getAllExternalContactsOrganizationFunc
	getExternalContactsOrganizationIdByNameAttr getExternalContactsOrganizationIdByNameFunc
	getExternalContactsOrganizationByIdAttr     getExternalContactsOrganizationByIdFunc
	updateExternalContactsOrganizationAttr      updateExternalContactsOrganizationFunc
	deleteExternalContactsOrganizationAttr      deleteExternalContactsOrganizationFunc
	externalOrganizationCache                   rc.CacheInterface[platformclientv2.Externalorganization]
}

// newExternalContactsOrganizationProxy initializes the external contacts organization proxy with all of the data needed to communicate with Genesys Cloud
func newExternalContactsOrganizationProxy(clientConfig *platformclientv2.Configuration) *externalContactsOrganizationProxy {
	api := platformclientv2.NewExternalContactsApiWithConfig(clientConfig)
	externalOrganizationCache := rc.NewResourceCache[platformclientv2.Externalorganization]()
	return &externalContactsOrganizationProxy{
		clientConfig:                                clientConfig,
		externalContactsApi:                         api,
		createExternalContactsOrganizationAttr:      createExternalContactsOrganizationFn,
		getAllExternalContactsOrganizationAttr:      getAllExternalContactsOrganizationFn,
		getExternalContactsOrganizationIdByNameAttr: getExternalContactsOrganizationIdByNameFn,
		getExternalContactsOrganizationByIdAttr:     getExternalContactsOrganizationByIdFn,
		updateExternalContactsOrganizationAttr:      updateExternalContactsOrganizationFn,
		deleteExternalContactsOrganizationAttr:      deleteExternalContactsOrganizationFn,
		externalOrganizationCache:                   externalOrganizationCache,
	}
}

// getExternalContactsOrganizationProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getExternalContactsOrganizationProxy(clientConfig *platformclientv2.Configuration) *externalContactsOrganizationProxy {
	if internalProxy == nil {
		internalProxy = newExternalContactsOrganizationProxy(clientConfig)
	}

	return internalProxy
}

// createExternalContactsOrganization creates a Genesys Cloud external contacts organization
func (p *externalContactsOrganizationProxy) createExternalContactsOrganization(ctx context.Context, externalContactsOrganization *platformclientv2.Externalorganization) (*platformclientv2.Externalorganization, *platformclientv2.APIResponse, error) {
	return p.createExternalContactsOrganizationAttr(ctx, p, externalContactsOrganization)
}

// getExternalContactsOrganization retrieves all Genesys Cloud external contacts organization
func (p *externalContactsOrganizationProxy) getAllExternalContactsOrganization(ctx context.Context, query string) (*[]platformclientv2.Externalorganization, *platformclientv2.APIResponse, error) {
	return p.getAllExternalContactsOrganizationAttr(ctx, p, query)
}

// getExternalContactsOrganizationIdByName returns a single Genesys Cloud external contacts organization by a name
func (p *externalContactsOrganizationProxy) getExternalContactsOrganizationIdByName(ctx context.Context, name string) (id string, retryable bool, apiResponse *platformclientv2.APIResponse, err error) {
	return p.getExternalContactsOrganizationIdByNameAttr(ctx, p, name)
}

// getExternalContactsOrganizationById returns a single Genesys Cloud external contacts organization by Id
func (p *externalContactsOrganizationProxy) getExternalContactsOrganizationById(ctx context.Context, id string) (externalContactsOrganization *platformclientv2.Externalorganization, apiResponse *platformclientv2.APIResponse, err error) {
	return p.getExternalContactsOrganizationByIdAttr(ctx, p, id)
}

// updateExternalContactsOrganization updates a Genesys Cloud external contacts organization
func (p *externalContactsOrganizationProxy) updateExternalContactsOrganization(ctx context.Context, id string, externalContactsOrganization *platformclientv2.Externalorganization) (*platformclientv2.Externalorganization, *platformclientv2.APIResponse, error) {
	return p.updateExternalContactsOrganizationAttr(ctx, p, id, externalContactsOrganization)
}

// deleteExternalContactsOrganization deletes a Genesys Cloud external contacts organization by Id
func (p *externalContactsOrganizationProxy) deleteExternalContactsOrganization(ctx context.Context, id string) (apiResponse *platformclientv2.APIResponse, err error) {
	return p.deleteExternalContactsOrganizationAttr(ctx, p, id)
}

// createExternalContactsOrganizationFn is an implementation function for creating a Genesys Cloud external contacts organization
func createExternalContactsOrganizationFn(ctx context.Context, p *externalContactsOrganizationProxy, externalContactsOrganization *platformclientv2.Externalorganization) (*platformclientv2.Externalorganization, *platformclientv2.APIResponse, error) {
	return p.externalContactsApi.PostExternalcontactsOrganizations(*externalContactsOrganization)
}

// getAllExternalContactsOrganizationFn is the implementation for retrieving all external contacts organization in Genesys Cloud
func getAllExternalContactsOrganizationFn(ctx context.Context, p *externalContactsOrganizationProxy, query string) (*[]platformclientv2.Externalorganization, *platformclientv2.APIResponse, error) {
	var allExternalOrganizations []platformclientv2.Externalorganization
	const pageSize = 100

	externalOrganizations, response, err := p.externalContactsApi.GetExternalcontactsOrganizations(pageSize, 1, query, nil, "", nil, true)
	if err != nil {
		return nil, response, fmt.Errorf("failed to get external organization: %v", err)
	}
	if externalOrganizations.Entities == nil || len(*externalOrganizations.Entities) == 0 {
		return &allExternalOrganizations, response, nil
	}
	allExternalOrganizations = append(allExternalOrganizations, *externalOrganizations.Entities...)

	for pageNum := 2; pageNum <= *externalOrganizations.PageCount; pageNum++ {
		externalOrganizations, response, err := p.externalContactsApi.GetExternalcontactsOrganizations(pageSize, pageNum, query, nil, "", nil, true)
		if err != nil {
			return nil, response, fmt.Errorf("failed to get external organization: %v", err)
		}

		if externalOrganizations.Entities == nil || len(*externalOrganizations.Entities) == 0 {
			break
		}

		allExternalOrganizations = append(allExternalOrganizations, *externalOrganizations.Entities...)

	}
	// Cache the External Contacts resource into the p.externalContactsCache for later use
	for _, externalOrganization := range allExternalOrganizations {
		if externalOrganization.Id == nil {
			continue
		}
		rc.SetCache(p.externalOrganizationCache, *externalOrganization.Id, externalOrganization)
	}

	return &allExternalOrganizations, response, nil
}

// getExternalContactsOrganizationIdByNameFn is an implementation of the function to get a Genesys Cloud external contacts organization by name
func getExternalContactsOrganizationIdByNameFn(ctx context.Context, p *externalContactsOrganizationProxy, name string) (id string, retryable bool, apiResponse *platformclientv2.APIResponse, err error) {

	externalOrganizations, response, err := p.getAllExternalContactsOrganization(ctx, name)
	if err != nil {
		return "", false, response, err
	}

	if externalOrganizations == nil || len(*externalOrganizations) == 0 {
		return "", true, response, fmt.Errorf("no external contacts organization found with name %s", name)
	}

	var externalOrganization platformclientv2.Externalorganization
	for _, externalOrganizationSdk := range *externalOrganizations {
		if *externalOrganizationSdk.Name == name {
			log.Printf("Retrieved the external contacts organization id %s by name %s", *externalOrganizationSdk.Id, name)
			externalOrganization = externalOrganizationSdk
			return *externalOrganization.Id, false, response, nil
		}
	}

	return "", true, response, fmt.Errorf("unable to find external contacts organization with name %s", name)
}

// getExternalContactsOrganizationByIdFn is an implementation of the function to get a Genesys Cloud external contacts organization by Id
func getExternalContactsOrganizationByIdFn(ctx context.Context, p *externalContactsOrganizationProxy, id string) (externalContactsOrganization *platformclientv2.Externalorganization, apiResponse *platformclientv2.APIResponse, err error) {
	if externalOrganization := rc.GetCacheItem(p.externalOrganizationCache, id); externalOrganization != nil {
		return externalOrganization, nil, nil
	}
	return p.externalContactsApi.GetExternalcontactsOrganization(id, []string{}, false)
}

// updateExternalContactsOrganizationFn is an implementation of the function to update a Genesys Cloud external contacts organization
func updateExternalContactsOrganizationFn(ctx context.Context, p *externalContactsOrganizationProxy, id string, externalContactsOrganization *platformclientv2.Externalorganization) (*platformclientv2.Externalorganization, *platformclientv2.APIResponse, error) {
	return p.externalContactsApi.PutExternalcontactsOrganization(id, *externalContactsOrganization)
}

// deleteExternalContactsOrganizationFn is an implementation function for deleting a Genesys Cloud external contacts organization
func deleteExternalContactsOrganizationFn(ctx context.Context, p *externalContactsOrganizationProxy, id string) (apiResponse *platformclientv2.APIResponse, err error) {
	_, response, err := p.externalContactsApi.DeleteExternalcontactsOrganization(id)
	return response, err
}
