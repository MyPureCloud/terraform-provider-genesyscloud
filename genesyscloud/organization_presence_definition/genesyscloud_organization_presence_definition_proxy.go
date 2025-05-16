package organization_presence_definition

import (
	"context"

	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
)

/*
The genesyscloud_organization_presence_definition_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *organizationPresenceDefinitionProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createOrganizationPresenceDefinitionFunc func(ctx context.Context, p *organizationPresenceDefinitionProxy, organizationPresenceDefinition *platformclientv2.Organizationpresencedefinition) (*platformclientv2.Organizationpresencedefinition, *platformclientv2.APIResponse, error)
type getAllOrganizationPresenceDefinitionFunc func(ctx context.Context, p *organizationPresenceDefinitionProxy) (*[]platformclientv2.Organizationpresencedefinition, *platformclientv2.APIResponse, error)
type getOrganizationPresenceDefinitionByIdFunc func(ctx context.Context, p *organizationPresenceDefinitionProxy, id string) (*platformclientv2.Organizationpresencedefinition, *platformclientv2.APIResponse, error)
type updateOrganizationPresenceDefinitionFunc func(ctx context.Context, p *organizationPresenceDefinitionProxy, id string, organizationPresenceDefinition *platformclientv2.Organizationpresencedefinition) (*platformclientv2.Organizationpresencedefinition, *platformclientv2.APIResponse, error)
type deleteOrganizationPresenceDefinitionFunc func(ctx context.Context, p *organizationPresenceDefinitionProxy, id string) (*platformclientv2.APIResponse, error)

// organizationPresenceDefinitionProxy contains all of the methods that call genesys cloud APIs.
type organizationPresenceDefinitionProxy struct {
	clientConfig                              *platformclientv2.Configuration
	presenceApi                               *platformclientv2.PresenceApi
	createOrganizationPresenceDefinitionAttr  createOrganizationPresenceDefinitionFunc
	getAllOrganizationPresenceDefinitionAttr  getAllOrganizationPresenceDefinitionFunc
	getOrganizationPresenceDefinitionByIdAttr getOrganizationPresenceDefinitionByIdFunc
	updateOrganizationPresenceDefinitionAttr  updateOrganizationPresenceDefinitionFunc
	deleteOrganizationPresenceDefinitionAttr  deleteOrganizationPresenceDefinitionFunc
}

// newOrganizationPresenceDefinitionProxy initializes the organization presence definition proxy with all of the data needed to communicate with Genesys Cloud
func newOrganizationPresenceDefinitionProxy(clientConfig *platformclientv2.Configuration) *organizationPresenceDefinitionProxy {
	api := platformclientv2.NewPresenceApiWithConfig(clientConfig)
	return &organizationPresenceDefinitionProxy{
		clientConfig:                              clientConfig,
		presenceApi:                               api,
		createOrganizationPresenceDefinitionAttr:  createOrganizationPresenceDefinitionFn,
		getAllOrganizationPresenceDefinitionAttr:  getAllOrganizationPresenceDefinitionFn,
		getOrganizationPresenceDefinitionByIdAttr: getOrganizationPresenceDefinitionByIdFn,
		updateOrganizationPresenceDefinitionAttr:  updateOrganizationPresenceDefinitionFn,
		deleteOrganizationPresenceDefinitionAttr:  deleteOrganizationPresenceDefinitionFn,
	}
}

// getOrganizationPresenceDefinitionProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getOrganizationPresenceDefinitionProxy(clientConfig *platformclientv2.Configuration) *organizationPresenceDefinitionProxy {
	if internalProxy == nil {
		internalProxy = newOrganizationPresenceDefinitionProxy(clientConfig)
	}

	return internalProxy
}

// createOrganizationPresenceDefinition creates a Genesys Cloud organization presence definition
func (p *organizationPresenceDefinitionProxy) createOrganizationPresenceDefinition(ctx context.Context, organizationPresenceDefinition *platformclientv2.Organizationpresencedefinition) (*platformclientv2.Organizationpresencedefinition, *platformclientv2.APIResponse, error) {
	return p.createOrganizationPresenceDefinitionAttr(ctx, p, organizationPresenceDefinition)
}

// getOrganizationPresenceDefinition retrieves all Genesys Cloud organization presence definition
func (p *organizationPresenceDefinitionProxy) getAllOrganizationPresenceDefinition(ctx context.Context) (*[]platformclientv2.Organizationpresencedefinition, *platformclientv2.APIResponse, error) {
	return p.getAllOrganizationPresenceDefinitionAttr(ctx, p)
}

// getOrganizationPresenceDefinitionById returns a single Genesys Cloud organization presence definition by Id
func (p *organizationPresenceDefinitionProxy) getOrganizationPresenceDefinitionById(ctx context.Context, id string) (*platformclientv2.Organizationpresencedefinition, *platformclientv2.APIResponse, error) {
	return p.getOrganizationPresenceDefinitionByIdAttr(ctx, p, id)
}

// updateOrganizationPresenceDefinition updates a Genesys Cloud organization presence definition
func (p *organizationPresenceDefinitionProxy) updateOrganizationPresenceDefinition(ctx context.Context, id string, organizationPresenceDefinition *platformclientv2.Organizationpresencedefinition) (*platformclientv2.Organizationpresencedefinition, *platformclientv2.APIResponse, error) {
	return p.updateOrganizationPresenceDefinitionAttr(ctx, p, id, organizationPresenceDefinition)
}

// deleteOrganizationPresenceDefinition deletes a Genesys Cloud organization presence definition by Id
func (p *organizationPresenceDefinitionProxy) deleteOrganizationPresenceDefinition(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	return p.deleteOrganizationPresenceDefinitionAttr(ctx, p, id)
}

// createOrganizationPresenceDefinitionFn is an implementation function for creating a Genesys Cloud organization presence definition
func createOrganizationPresenceDefinitionFn(ctx context.Context, p *organizationPresenceDefinitionProxy, organizationPresenceDefinition *platformclientv2.Organizationpresencedefinition) (*platformclientv2.Organizationpresencedefinition, *platformclientv2.APIResponse, error) {
	return p.presenceApi.PostPresenceDefinitions(*organizationPresenceDefinition)
}

// getAllOrganizationPresenceDefinitionFn is the implementation for retrieving all organization presence definition in Genesys Cloud
func getAllOrganizationPresenceDefinitionFn(ctx context.Context, p *organizationPresenceDefinitionProxy) (*[]platformclientv2.Organizationpresencedefinition, *platformclientv2.APIResponse, error) {
	var allOrganizationPresenceDefinitions []platformclientv2.Organizationpresencedefinition

	// Get active organization presence definitions
	activeOrganizationPresenceDefinitions, resp, err := p.presenceApi.GetPresenceDefinitions("FALSE", []string{}, "")
	if err != nil {
		return nil, resp, err
	}
	allOrganizationPresenceDefinitions = append(allOrganizationPresenceDefinitions, *activeOrganizationPresenceDefinitions.Entities...)

	// By default the api only returns active definitions, so get inactive organization presence definitions
	inactiveOrganizationPresenceDefinitions, resp, err := p.presenceApi.GetPresenceDefinitions("TRUE", []string{}, "")
	if err != nil {
		return nil, resp, err
	}
	allOrganizationPresenceDefinitions = append(allOrganizationPresenceDefinitions, *inactiveOrganizationPresenceDefinitions.Entities...)

	return &allOrganizationPresenceDefinitions, resp, nil
}

// getOrganizationPresenceDefinitionByIdFn is an implementation of the function to get a Genesys Cloud organization presence definition by Id
func getOrganizationPresenceDefinitionByIdFn(ctx context.Context, p *organizationPresenceDefinitionProxy, id string) (*platformclientv2.Organizationpresencedefinition, *platformclientv2.APIResponse, error) {
	return p.presenceApi.GetPresenceDefinition(id, "")
}

// updateOrganizationPresenceDefinitionFn is an implementation of the function to update a Genesys Cloud organization presence definition
func updateOrganizationPresenceDefinitionFn(ctx context.Context, p *organizationPresenceDefinitionProxy, id string, organizationPresenceDefinition *platformclientv2.Organizationpresencedefinition) (*platformclientv2.Organizationpresencedefinition, *platformclientv2.APIResponse, error) {
	return p.presenceApi.PutPresenceDefinition(id, *organizationPresenceDefinition)
}

// deleteOrganizationPresenceDefinitionFn is an implementation function for deleting a Genesys Cloud organization presence definition
func deleteOrganizationPresenceDefinitionFn(ctx context.Context, p *organizationPresenceDefinitionProxy, id string) (*platformclientv2.APIResponse, error) {
	return p.presenceApi.DeletePresenceDefinition(id)
}
