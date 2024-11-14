package external_contacts_organization

import (
	"context"
	"fmt"

	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

/*
The genesyscloud_external_contacts_organization_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *externalContactsOrganizationProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createExternalContactsOrganizationFunc func(ctx context.Context, p *externalContactsOrganizationProxy, externalOrganization *platformclientv2.Externalorganization) (*platformclientv2.Externalorganization, error)
type getAllExternalContactsOrganizationFunc func(ctx context.Context, p *externalContactsOrganizationProxy) (*[]platformclientv2.Externalorganization, error)
type getExternalContactsOrganizationIdByNameFunc func(ctx context.Context, p *externalContactsOrganizationProxy, name string) (id string, retryable bool, err error)
type getExternalContactsOrganizationByIdFunc func(ctx context.Context, p *externalContactsOrganizationProxy, id string) (externalOrganization *platformclientv2.Externalorganization, responseCode int, err error)
type updateExternalContactsOrganizationFunc func(ctx context.Context, p *externalContactsOrganizationProxy, id string, externalOrganization *platformclientv2.Externalorganization) (*platformclientv2.Externalorganization, error)
type deleteExternalContactsOrganizationFunc func(ctx context.Context, p *externalContactsOrganizationProxy, id string) (responseCode int, err error)

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
}

// newExternalContactsOrganizationProxy initializes the external contacts organization proxy with all of the data needed to communicate with Genesys Cloud
func newExternalContactsOrganizationProxy(clientConfig *platformclientv2.Configuration) *externalContactsOrganizationProxy {
	api := platformclientv2.NewExternalContactsApiWithConfig(clientConfig)
	return &externalContactsOrganizationProxy{
		clientConfig:                                clientConfig,
		externalContactsApi:                         api,
		createExternalContactsOrganizationAttr:      createExternalContactsOrganizationFn,
		getAllExternalContactsOrganizationAttr:      getAllExternalContactsOrganizationFn,
		getExternalContactsOrganizationIdByNameAttr: getExternalContactsOrganizationIdByNameFn,
		getExternalContactsOrganizationByIdAttr:     getExternalContactsOrganizationByIdFn,
		updateExternalContactsOrganizationAttr:      updateExternalContactsOrganizationFn,
		deleteExternalContactsOrganizationAttr:      deleteExternalContactsOrganizationFn,
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
func (p *externalContactsOrganizationProxy) createExternalContactsOrganization(ctx context.Context, externalContactsOrganization *platformclientv2.Externalorganization) (*platformclientv2.Externalorganization, error) {
	return p.createExternalContactsOrganizationAttr(ctx, p, externalContactsOrganization)
}

// getExternalContactsOrganization retrieves all Genesys Cloud external contacts organization
func (p *externalContactsOrganizationProxy) getAllExternalContactsOrganization(ctx context.Context) (*[]platformclientv2.Externalorganization, error) {
	return p.getAllExternalContactsOrganizationAttr(ctx, p)
}

// getExternalContactsOrganizationIdByName returns a single Genesys Cloud external contacts organization by a name
func (p *externalContactsOrganizationProxy) getExternalContactsOrganizationIdByName(ctx context.Context, name string) (id string, retryable bool, err error) {
	return p.getExternalContactsOrganizationIdByNameAttr(ctx, p, name)
}

// getExternalContactsOrganizationById returns a single Genesys Cloud external contacts organization by Id
func (p *externalContactsOrganizationProxy) getExternalContactsOrganizationById(ctx context.Context, id string) (externalContactsOrganization *platformclientv2.Externalorganization, statusCode int, err error) {
	return p.getExternalContactsOrganizationByIdAttr(ctx, p, id)
}

// updateExternalContactsOrganization updates a Genesys Cloud external contacts organization
func (p *externalContactsOrganizationProxy) updateExternalContactsOrganization(ctx context.Context, id string, externalContactsOrganization *platformclientv2.Externalorganization) (*platformclientv2.Externalorganization, error) {
	return p.updateExternalContactsOrganizationAttr(ctx, p, id, externalContactsOrganization)
}

// deleteExternalContactsOrganization deletes a Genesys Cloud external contacts organization by Id
func (p *externalContactsOrganizationProxy) deleteExternalContactsOrganization(ctx context.Context, id string) (statusCode int, err error) {
	return p.deleteExternalContactsOrganizationAttr(ctx, p, id)
}

// createExternalContactsOrganizationFn is an implementation function for creating a Genesys Cloud external contacts organization
func createExternalContactsOrganizationFn(ctx context.Context, p *externalContactsOrganizationProxy, externalContactsOrganization *platformclientv2.Externalorganization) (*platformclientv2.Externalorganization, error) {
	externalOrganization, _, err := p.externalContactsApi.PostExternalcontactsOrganizations(*externalContactsOrganization)
	if err != nil {
		return nil, fmt.Errorf("Failed to create external contacts organization: %s", err)
	}

	return externalOrganization, nil
}

// getAllExternalContactsOrganizationFn is the implementation for retrieving all external contacts organization in Genesys Cloud
func getAllExternalContactsOrganizationFn(ctx context.Context, p *externalContactsOrganizationProxy) (*[]platformclientv2.Externalorganization, error) {
	var allExternalOrganizations []platformclientv2.Externalorganization
	const pageSize = 100
	/*
		externalOrganizations, _, err := p.externalContactsApi.GetExternalcontactsOrganizations()
		if err != nil {
			return nil, fmt.Errorf("Failed to get external organization: %v", err)
		}
		if externalOrganizations.Entities == nil || len(*externalOrganizations.Entities) == 0 {
			return &allExternalOrganizations, nil
		}
		for _, externalOrganization := range *externalOrganizations.Entities {
			allExternalOrganizations = append(allExternalOrganizations, externalOrganization)
		}

		for pageNum := 2; pageNum <= *externalOrganizations.PageCount; pageNum++ {
			externalOrganizations, _, err := p.externalContactsApi.GetExternalcontactsOrganizations()
			if err != nil {
				return nil, fmt.Errorf("Failed to get external organization: %v", err)
			}

			if externalOrganizations.Entities == nil || len(*externalOrganizations.Entities) == 0 {
				break
			}

			for _, externalOrganization := range *externalOrganizations.Entities {
				allExternalOrganizations = append(allExternalOrganizations, externalOrganization)
			}
		}
	*/
	return &allExternalOrganizations, nil
}

// getExternalContactsOrganizationIdByNameFn is an implementation of the function to get a Genesys Cloud external contacts organization by name
func getExternalContactsOrganizationIdByNameFn(ctx context.Context, p *externalContactsOrganizationProxy, name string) (id string, retryable bool, err error) {
	/*
		externalOrganizations, _, err := p.externalContactsApi.GetExternalcontactsOrganizations()
		if err != nil {
			return "", false, err
		}

		if externalOrganizations.Entities == nil || len(*externalOrganizations.Entities) == 0 {
			return "", true, fmt.Errorf("No external contacts organization found with name %s", name)
		}

		var externalOrganization platformclientv2.Externalorganization
		for _, externalOrganizationSdk := range *externalOrganizations.Entities {
			if *externalOrganization.Name == name {
				log.Printf("Retrieved the external contacts organization id %s by name %s", *externalOrganizationSdk.Id, name)
				externalOrganization = externalOrganizationSdk
				return *externalOrganization.Id, false, nil
			}
		}
	*/
	return "", false, fmt.Errorf("Unable to find external contacts organization with name %s", name)
}

// getExternalContactsOrganizationByIdFn is an implementation of the function to get a Genesys Cloud external contacts organization by Id
func getExternalContactsOrganizationByIdFn(ctx context.Context, p *externalContactsOrganizationProxy, id string) (externalContactsOrganization *platformclientv2.Externalorganization, statusCode int, err error) {
	externalOrganization, resp, err := p.externalContactsApi.GetExternalcontactsOrganization(id, []string{}, false)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("Failed to retrieve external contacts organization by id %s: %s", id, err)
	}

	return externalOrganization, resp.StatusCode, nil
}

// updateExternalContactsOrganizationFn is an implementation of the function to update a Genesys Cloud external contacts organization
func updateExternalContactsOrganizationFn(ctx context.Context, p *externalContactsOrganizationProxy, id string, externalContactsOrganization *platformclientv2.Externalorganization) (*platformclientv2.Externalorganization, error) {
	externalOrganization, _, err := p.externalContactsApi.PutExternalcontactsOrganization(id, *externalContactsOrganization)
	if err != nil {
		return nil, fmt.Errorf("Failed to update external contacts organization: %s", err)
	}
	return externalOrganization, nil
}

// deleteExternalContactsOrganizationFn is an implementation function for deleting a Genesys Cloud external contacts organization
func deleteExternalContactsOrganizationFn(ctx context.Context, p *externalContactsOrganizationProxy, id string) (statusCode int, err error) {
	_, resp, err := p.externalContactsApi.DeleteExternalcontactsOrganization(id)
	if err != nil {
		return resp.StatusCode, fmt.Errorf("Failed to delete external contacts organization: %s", err)
	}

	return resp.StatusCode, nil
}
